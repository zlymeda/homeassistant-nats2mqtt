package executor

import (
	"github.com/shopspring/decimal"
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddNumber(e *entity.Number) error {
	stateObs := observable.NewMapped(e.State, func(d decimal.Decimal) string {
		return d.String()
	})

	if err := s.monitorDecimalCmd(e.Meta.Current(), e.Set, entity.CommandTopic); err != nil {
		return err
	}

	s.register(entity.Entity{
		Meta: observable.NewMapped(e.Meta, func(i entity.NumberMeta) entity.Metadata {
			i.ExtraTopics.AddIfNotNull(e.Set, entity.CommandTopic)
			return i
		}),
		State:      stateObs,
		Attributes: e.Attributes,
	})

	return nil
}
