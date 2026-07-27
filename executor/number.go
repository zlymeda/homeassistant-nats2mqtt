package executor

import (
	"github.com/shopspring/decimal"
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddNumber(e *entity.Number) error {
	// Build a state observable that preserves mutability when the source is mutable.
	// This is needed for MqttOnly entities which require Mutable[string] to receive updates.
	var stateObs observable.Observable[string]
	if mutable, ok := e.State.(observable.Mutable[decimal.Decimal]); ok {
		stateObs = observable.NewMutableMapped(mutable,
			func(d decimal.Decimal) string { return d.String() },
			func(s string) decimal.Decimal { return decimal.RequireFromString(s) })
	} else {
		stateObs = observable.NewMapped(e.State, func(d decimal.Decimal) string {
			return d.String()
		})
	}

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
