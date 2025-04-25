package entity

import (
	"github.com/shopspring/decimal"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

var _ Metadata = NumberMeta{}

type NumberMeta struct {
	Meta

	Mode string
	Min  decimal.Decimal
	Step decimal.Decimal
	Max  decimal.Decimal

	ExtraTopics ExtraTopicsList
}

func (n NumberMeta) GetId() string {
	return n.Id
}

func (n NumberMeta) GetName() string {
	return n.Name
}

func (n NumberMeta) GetExtraTopics() []string {
	return n.ExtraTopics
}

func (n NumberMeta) ToHaDiscovery(dev Device) map[string]any {
	result := n.Meta.ToHaDiscovery(dev, "number")

	if n.Mode != "" {
		result["mode"] = n.Mode
	}
	if !n.Min.IsZero() {
		result["min"] = n.Min.String()
	}
	if !n.Step.IsZero() {
		result["step"] = n.Step.String()
	}
	if !n.Max.IsZero() {
		result["max"] = n.Max.String()
	}

	return result
}

type Number struct {
	Meta       observable.Observable[NumberMeta]
	State      observable.Observable[decimal.Decimal]
	Attributes observable.Observable[Attrs]

	Set func(decimal.Decimal) error
}
