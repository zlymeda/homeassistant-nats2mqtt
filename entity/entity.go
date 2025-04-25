package entity

import (
	"fmt"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

const (
	CommandTopic = `cmd`
)

type Attrs map[string]any

type Metadata interface {
	GetId() string
	GetName() string
	GetExtraTopics() []string
	ToHaDiscovery(dev Device) map[string]any
}

type Entity struct {
	Meta       observable.Observable[Metadata]
	State      observable.Observable[string]
	Attributes observable.Observable[Attrs]
}

type Meta struct {
	Id             string
	Name           string
	DeviceClass    DeviceClass
	StateClass     StateClass
	Icon           string
	Unit           string
	EntityCategory Category
}

func (m Meta) ToHaDiscovery(dev Device, platform string) map[string]any {
	result := map[string]any{
		"p":       platform,
		"name":    m.Name,
		"obj_id":  fmt.Sprintf("%s_%s", dev.Name, m.Id),
		"uniq_id": fmt.Sprintf("%s_%s", dev.Id, m.Id),
	}

	if m.DeviceClass != "" {
		result["dev_cla"] = m.DeviceClass
	}
	if m.Icon != "" {
		result["ic"] = m.Icon
	}
	if m.Unit != "" {
		result["unit_of_meas"] = m.Unit
	}
	if m.StateClass != "" {
		result["stat_cla"] = m.StateClass
	}
	if m.EntityCategory != "" {
		result["ent_cat"] = m.EntityCategory
	}

	return result
}
