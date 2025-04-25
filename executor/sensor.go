package executor

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddSensor(e *entity.Sensor) {
	s.register(entity.Entity{
		Meta: observable.NewMapped(e.Meta, func(i entity.Meta) entity.Metadata {
			return entity.SensorMeta{
				Meta: i,
			}
		}),
		State:      e.State,
		Attributes: e.Attributes,
	})
}
