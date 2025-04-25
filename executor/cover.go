package executor

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddCover(e *entity.Cover) error {
	coverMeta := observable.NewMapped(e.Meta, func(i entity.Meta) entity.Metadata {
		return entity.CoverMeta{
			Meta:         i,
			SupportStop:  e.Stop != nil,
			SupportOpen:  e.Open != nil,
			SupportClose: e.Close != nil,
		}
	})

	s.register(entity.Entity{
		Meta: coverMeta,
		State: observable.NewMapped(e.State, func(i entity.CoverState) string {
			return string(i)
		}),
		Attributes: e.Attributes,
	})

	commands := map[string]func() error{
		"OPEN":  e.Open,
		"CLOSE": e.Close,
		"STOP":  e.Stop,
	}

	meta := coverMeta.Current()
	cmdTopic := s.fullTopic(meta, entity.CommandTopic)
	if err := s.monitorCommandsOn(cmdTopic, createCallback(commands)); err != nil {
		return err
	}

	return nil
}
