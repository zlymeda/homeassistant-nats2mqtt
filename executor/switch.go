package executor

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddSwitch(e *entity.Switch) error {
	switchMeta := observable.NewMapped(e.Meta, func(i entity.Meta) entity.Metadata {
		return entity.SwitchMeta{
			Meta: i,
		}
	})

	s.register(entity.Entity{
		Meta: switchMeta,
		State: observable.NewMapped(e.State, func(i entity.SwitchState) string {
			return string(i)
		}),
		Attributes: e.Attributes,
	})

	commands := map[string]func() error{
		"ON":  e.TurnOn,
		"OFF": e.TurnOff,
	}

	meta := switchMeta.Current()

	cmdTopic := s.fullTopic(meta, entity.CommandTopic)
	if err := s.monitorCommandsOn(cmdTopic, createCallback(commands)); err != nil {
		return err
	}

	return nil
}
