package entity

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

var _ Metadata = SwitchMeta{}

type SwitchMeta struct {
	Meta
}

func (s SwitchMeta) GetId() string {
	return s.Id
}

func (s SwitchMeta) GetName() string {
	return s.Name
}

func (s SwitchMeta) GetExtraTopics() []string {
	return []string{CommandTopic}
}

func (s SwitchMeta) ToHaDiscovery(dev Device) map[string]any {
	return s.Meta.ToHaDiscovery(dev, "switch")
}

type SwitchState string

const (
	SwitchStateOn  SwitchState = "ON"
	SwitchStateOff SwitchState = "OFF"
)

type Switch struct {
	Meta       observable.Observable[Meta]
	State      observable.Observable[SwitchState]
	Attributes observable.Observable[Attrs]

	TurnOn  func() error
	TurnOff func() error
}
