package entity

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

var _ Metadata = ButtonMeta{}

type ButtonMeta struct {
	Meta
}

func (s ButtonMeta) GetId() string {
	return s.Id
}

func (s ButtonMeta) GetName() string {
	return s.Name
}

func (s ButtonMeta) GetExtraTopics() []string {
	return []string{CommandTopic}
}

func (s ButtonMeta) ToHaDiscovery(dev Device) map[string]any {
	return s.Meta.ToHaDiscovery(dev, "button")
}

type Button struct {
	Meta       observable.Observable[Meta]
	Attributes observable.Observable[Attrs]

	Press func() error
}
