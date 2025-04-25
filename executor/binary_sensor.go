package executor

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddBinarySensor(e *entity.BinarySensor) {
	s.register(entity.Entity{
		Meta: observable.NewMapped(e.Meta, func(i entity.Meta) entity.Metadata {
			return entity.BinarySensorMeta{
				Meta: i,
			}
		}),
		State: observable.NewMapped(e.State, func(i entity.BinarySensorState) string {
			return string(i)
		}),
		Attributes: e.Attributes,
	})
}

func (s *EntityRegistry) AddBinarySensors(sensors []*entity.BinarySensor) {
	for _, e := range sensors {
		s.AddBinarySensor(e)
	}
}
