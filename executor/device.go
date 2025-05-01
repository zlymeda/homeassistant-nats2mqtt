package executor

import "github.com/zlymeda/homeassistant-nats2mqtt/entity"

type Discovery struct {
	Device entity.Device `json:"dev"`
	Origin Origin        `json:"o"`

	Entities map[string]map[string]any `json:"cmps"`
}

type Origin struct {
	Name            string `json:"name"`
	SoftwareVersion string `json:"sw"`
}

func (s *Service) AddDevice(device entity.Device) *EntityRegistry {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.devices[device.Id] = &EntityRegistry{
		ctx:         s.ctx,
		nc:          s.nc,
		device:      device,
		topicPrefix: Topic(s.origin.Name, device.Id),
		origin:      s.origin,
		stateUpdated: func() {
			s.stateUpdated()
		},
	}

	s.stateUpdated()

	return s.devices[device.Id]
}

func (s *Service) stateUpdated() {
	select {
	case s.updates <- struct{}{}:
	default:
	}
}
