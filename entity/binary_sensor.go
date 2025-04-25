package entity

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

var _ Metadata = BinarySensorMeta{}

type BinarySensorMeta struct {
	Meta
}

func (b BinarySensorMeta) GetId() string {
	return b.Id
}

func (b BinarySensorMeta) GetName() string {
	return b.Name
}

func (b BinarySensorMeta) GetExtraTopics() []string {
	return nil
}

func (b BinarySensorMeta) ToHaDiscovery(dev Device) map[string]any {
	result := b.Meta.ToHaDiscovery(dev, "binary_sensor")
	return result
}

type BinarySensorState string

const (
	BinarySensorStateOn  BinarySensorState = "ON"
	BinarySensorStateOff BinarySensorState = "OFF"
)

type BinarySensor struct {
	Meta       observable.Observable[Meta]
	State      observable.Observable[BinarySensorState]
	Attributes observable.Observable[Attrs]
}
