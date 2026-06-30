package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

type EntityRegistry struct {
	ctx    context.Context
	nc     *nats.Conn
	device entity.Device

	mutex sync.Mutex

	entities []entity.Entity

	topicPrefix    string
	rawStatePrefix string

	publishState []func() error
	origin       Origin

	stateUpdated func()
}

func (s *EntityRegistry) PublishDiscovery() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	devId := s.device.Id

	entities := map[string]map[string]any{}
	for _, v := range s.entities {
		meta := v.Meta.Current()

		values := meta.ToHaDiscovery(s.device)

		if v.State != nil {
			values["stat_t"] = natsTopicToMqtt(s.stateTopic(meta))
		}
		if v.Attributes != nil {
			values["json_attr_t"] = natsTopicToMqtt(s.attrsTopic(meta))
		}

		for _, topic := range meta.GetExtraTopics() {
			values[topic+"_t"] = natsTopicToMqtt(s.fullTopic(meta, topic))
		}

		platform, _ := values["p"].(string)
		uniqueId, _ := values["uniq_id"].(string)

		path := fmt.Sprintf("%s.%s", platform, uniqueId)
		entities[path] = values
	}

	cfg, err := json.Marshal(Discovery{
		Device:   s.device,
		Origin:   s.origin,
		Entities: entities,
	})

	if err != nil {
		return fmt.Errorf("failed to marshal JSON device %s config: %w", devId, err)
	}

	// <discovery_prefix>/<component>/[<node_id>/]<object_id>/config
	configTopic := Topic("homeassistant", "device", s.origin.Name, devId, "config")
	err = s.nc.Publish(configTopic, cfg)
	if err != nil {
		return fmt.Errorf("failed to publish device %s config: %w", devId, err)
	}

	return nil
}

// PublishStates publishes the current state of all entities.
// Must be called while holding Service.mutex to protect the devices map,
// and acquires EntityRegistry.mutex internally to protect publishState.
func (s *EntityRegistry) PublishStates() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var errs error
	for _, publish := range s.publishState {
		errs = errors.Join(errs, publish())
	}

	return errs
}

func (s *EntityRegistry) monitorDiscovery(e entity.Entity) {
	changes := e.Meta.Changes()
	if changes == nil {
		return
	}

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return

			case <-changes:
				if err := s.PublishDiscovery(); err != nil {
					slog.Error("publishing discovery failed",
						slog.String("device.id", s.device.Id),
						slog.String("device.name", s.device.Name),
						slog.Any("error", err),
					)
				}
			}
		}
	}()
}

func (s *EntityRegistry) monitorState(e entity.Entity) {
	state := e.State
	if state == nil {
		return
	}

	meta := e.Meta

	monitorObservable(s, meta, state, func(state string) error {
		currentMeta := meta.Current()
		topic := s.stateTopic(currentMeta)
		err := s.nc.Publish(topic, []byte(state))
		if err != nil {
			return fmt.Errorf("error publishing new state for %s: %w", currentMeta.GetName(), err)
		}
		if err := s.publishRaw(currentMeta, "state", []byte(state)); err != nil {
			return fmt.Errorf("error publishing raw state for %s: %w", currentMeta.GetName(), err)
		}
		return nil
	})
}

func (s *EntityRegistry) monitorAttributes(e entity.Entity) {
	attributes := e.Attributes
	if attributes == nil {
		return
	}

	meta := e.Meta

	monitorObservable(s, meta, attributes, func(state entity.Attrs) error {
		currentMeta := meta.Current()
		topic := s.attrsTopic(currentMeta)
		res, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("error marshalling new attrs for %s: %w", currentMeta.GetName(), err)
		}

		err = s.nc.Publish(topic, res)
		if err != nil {
			return fmt.Errorf("error publishing new attrs for %s: %w", currentMeta.GetName(), err)
		}
		if err := s.publishRaw(currentMeta, "attrs", res); err != nil {
			return fmt.Errorf("error publishing raw attrs for %s: %w", currentMeta.GetName(), err)
		}

		return nil
	})
}

func (s *EntityRegistry) stateTopic(meta entity.Metadata) string {
	return s.fullTopic(meta, "state")
}

func (s *EntityRegistry) attrsTopic(meta entity.Metadata) string {
	return s.fullTopic(meta, "attrs")
}

func (s *EntityRegistry) fullTopic(meta entity.Metadata, subTopic string) string {
	platform, _ := meta.ToHaDiscovery(s.device)["p"].(string)
	return Topic(s.topicPrefix, platform, meta.GetId(), subTopic)
}

func (s *EntityRegistry) rawTopic(meta entity.Metadata, subTopic string) string {
	platform, _ := meta.ToHaDiscovery(s.device)["p"].(string)
	return Topic(s.rawStatePrefix, s.origin.Name, s.device.Id, platform, meta.GetId(), subTopic)
}

func (s *EntityRegistry) publishRaw(meta entity.Metadata, subTopic string, data []byte) error {
	if s.rawStatePrefix == "" {
		return nil
	}
	return s.nc.Publish(s.rawTopic(meta, subTopic), data)
}

func (s *EntityRegistry) register(e entity.Entity) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.entities = append(s.entities, e)

	s.monitorState(e)
	s.monitorAttributes(e)
	s.monitorDiscovery(e)
	s.stateUpdated()
}

func monitorObservable[T any](s *EntityRegistry, meta observable.Observable[entity.Metadata], obs observable.Observable[T], publish func(state T) error) {
	s.publishState = append(s.publishState, func() error {
		return publish(obs.Current())
	})

	changes := obs.Changes()
	if changes == nil {
		return
	}

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return

			case change := <-changes:
				currentMeta := meta.Current()
				err := publish(change)
				if err != nil {
					slog.Error("error publishing",
						slog.String("name", currentMeta.GetName()),
						slog.Any("value", change),
						slog.Any("err", err),
					)
				}
			}
		}
	}()
}

func natsTopicToMqtt(topic string) string {
	return strings.ReplaceAll(topic, ".", "/")
}
