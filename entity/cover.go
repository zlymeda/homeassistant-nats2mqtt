package entity

import "github.com/zlymeda/homeassistant-nats2mqtt/observable"

var _ Metadata = CoverMeta{}

type CoverMeta struct {
	Meta

	SupportStop  bool
	SupportOpen  bool
	SupportClose bool
}

func (c CoverMeta) GetId() string {
	return c.Id
}

func (c CoverMeta) GetName() string {
	return c.Name
}

func (c CoverMeta) GetExtraTopics() []string {
	return []string{CommandTopic}
}

func (c CoverMeta) ToHaDiscovery(dev Device) map[string]any {
	result := c.Meta.ToHaDiscovery(dev, "cover")

	if !c.SupportStop {
		result["pl_stop"] = nil
	}
	if !c.SupportOpen {
		result["pl_open"] = nil
	}
	if !c.SupportClose {
		result["pl_cls"] = nil
	}

	return result
}

type CoverState string

const (
	CoverStateOpen   CoverState = "open"
	CoverStateClosed CoverState = "closed"
)

type Cover struct {
	Meta       observable.Observable[Meta]
	State      observable.Observable[CoverState]
	Attributes observable.Observable[Attrs]

	Open  func() error
	Close func() error
	Stop  func() error
}
