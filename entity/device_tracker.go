package entity

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

var _ Metadata = DeviceTrackerMeta{}

type DeviceTrackerMeta struct {
	Meta
}

func (d DeviceTrackerMeta) GetId() string {
	return d.Id
}

func (d DeviceTrackerMeta) GetName() string {
	return d.Name
}

func (d DeviceTrackerMeta) GetExtraTopics() []string {
	return nil
}

func (d DeviceTrackerMeta) ToHaDiscovery(dev Device) map[string]any {
	return d.Meta.ToHaDiscovery(dev, "device_tracker")
}

type DeviceTracker struct {
	Meta       observable.Observable[Meta]
	State      observable.Observable[string]
	Attributes observable.Observable[Attrs]
}
