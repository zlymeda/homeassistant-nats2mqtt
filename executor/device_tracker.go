package executor

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddDeviceTracker(e *entity.DeviceTracker) {
	s.register(entity.Entity{
		Meta: observable.NewMapped(e.Meta, func(i entity.Meta) entity.Metadata {
			return entity.DeviceTrackerMeta{
				Meta: i,
			}
		}),
		State:      e.State,
		Attributes: e.Attributes,
	})
}
