package entity

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

var _ Metadata = SensorMeta{}

type SensorMeta struct {
	Meta
}

func (s SensorMeta) GetId() string {
	return s.Id
}

func (s SensorMeta) GetName() string {
	return s.Name
}

func (s SensorMeta) GetExtraTopics() []string {
	return nil
}

func (s SensorMeta) ToHaDiscovery(dev Device) map[string]any {
	return s.Meta.ToHaDiscovery(dev, "sensor")
}

type Sensor struct {
	Meta       observable.Observable[Meta]
	State      observable.Observable[string]
	Attributes observable.Observable[Attrs]
}
