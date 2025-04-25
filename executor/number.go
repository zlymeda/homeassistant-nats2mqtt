package executor

import (
	"github.com/shopspring/decimal"
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddNumber(e *entity.Number) error {
	s.register(entity.Entity{
		Meta: observable.NewMapped(e.Meta, func(i entity.NumberMeta) entity.Metadata {
			i.ExtraTopics.AddIfNotNull(e.Set, entity.CommandTopic)

			return i
		}),
		State: observable.NewMapped(e.State, func(i decimal.Decimal) string {
			return i.String()
		}),
		Attributes: e.Attributes,
	})

	return s.monitorDecimalCmd(e.Meta.Current(), e.Set, entity.CommandTopic)
}
