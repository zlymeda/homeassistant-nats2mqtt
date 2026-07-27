package executor

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddSelect(e *entity.Select) error {
	if err := s.monitorStringCmd(e.Meta.Current(), e.Set, entity.CommandTopic); err != nil {
		return err
	}

	s.register(entity.Entity{
		Meta: observable.NewMapped(e.Meta, func(i entity.SelectMeta) entity.Metadata {
			i.ExtraTopics.AddIfNotNull(e.Set, entity.CommandTopic)
			return i
		}),
		State:      e.State,
		Attributes: e.Attributes,
	})

	return nil
}
