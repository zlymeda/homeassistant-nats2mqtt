package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
	"log/slog"
	"strings"
	"sync"
)

type EntityRegistry struct {
	ctx    context.Context
	nc     *nats.Conn
	device entity.Device

	mutex sync.Mutex

	entities []entity.Entity

	topicPrefix string

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

		platform := values["p"].(string)
		uniqueId := values["uniq_id"].(string)

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

	meta := e.Meta.Current()
	topic := s.stateTopic(meta)

	monitorObservable(s, meta, state, func(state string) error {
		err := s.nc.Publish(topic, []byte(state))
		if err != nil {
			return fmt.Errorf("error publishing new state for %s: %w", meta.GetName(), err)
		}
		return nil
	})
}

func (s *EntityRegistry) monitorAttributes(e entity.Entity) {
	attributes := e.Attributes
	if attributes == nil {
		return
	}

	meta := e.Meta.Current()
	topic := s.attrsTopic(meta)

	monitorObservable(s, meta, attributes, func(state entity.Attrs) error {
		res, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("error marshalling new attrs for %s: %w", meta.GetName(), err)
		}

		err = s.nc.Publish(topic, res)
		if err != nil {
			return fmt.Errorf("error publishing new attrs for %s: %w", meta.GetName(), err)
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
	platform := meta.ToHaDiscovery(s.device)["p"].(string)
	return Topic(s.topicPrefix, platform, meta.GetId(), subTopic)
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

func monitorObservable[T any](s *EntityRegistry, meta entity.Metadata, observable observable.Observable[T], publish func(state T) error) {
	s.publishState = append(s.publishState, func() error {
		return publish(observable.Current())
	})

	changes := observable.Changes()
	if changes == nil {
		return
	}

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return

			case change := <-changes:
				err := publish(change)
				if err != nil {
					slog.Error("error publishing",
						slog.String("name", meta.GetName()),
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
