package executor

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddButton(e *entity.Button) error {
	buttonMeta := observable.NewMapped(e.Meta, func(i entity.Meta) entity.Metadata {
		return entity.ButtonMeta{
			Meta: i,
		}
	})

	s.register(entity.Entity{
		Meta:       buttonMeta,
		Attributes: e.Attributes,
	})

	meta := buttonMeta.Current()

	cmdTopic := s.fullTopic(meta, entity.CommandTopic)
	if err := s.monitorCommandsOn(cmdTopic, func(bytes []byte) error {
		return e.Press()
	}); err != nil {
		return err
	}

	return nil
}
