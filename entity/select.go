package entity

import "github.com/zlymeda/homeassistant-nats2mqtt/observable"

var _ Metadata = SelectMeta{}

type SelectMeta struct {
	Meta

	Options []string

	ExtraTopics ExtraTopicsList
}

func (s SelectMeta) GetId() string {
	return s.Id
}

func (s SelectMeta) GetName() string {
	return s.Name
}

func (s SelectMeta) GetExtraTopics() []string {
	return s.ExtraTopics
}

func (s SelectMeta) ToHaDiscovery(dev Device) map[string]any {
	result := s.Meta.ToHaDiscovery(dev, "select")

	if len(s.Options) > 0 {
		result["options"] = s.Options
	}

	return result
}

type Select struct {
	Meta       observable.Observable[SelectMeta]
	State      observable.Observable[string]
	Attributes observable.Observable[Attrs]

	Set func(string) error
}
