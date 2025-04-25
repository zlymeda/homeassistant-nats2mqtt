package entity

import (
	"github.com/shopspring/decimal"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

var _ Metadata = ClimateMeta{}

const (
	ClimateCurrentTemperatureTopic = `curr_temp`

	ClimateTemperatureStateTopic   = `temp_stat`
	ClimateTemperatureCommandTopic = `temp_cmd`

	ClimateModeStateTopic   = `mode_stat`
	ClimateModeCommandTopic = `mode_cmd`

	ClimatePresetModeStateTopic   = `pr_mode_stat`
	ClimatePresetModeCommandTopic = `pr_mode_cmd`
)

// ExtraTopicsList is a helper type that extends []string with an AddIfNotNull method
type ExtraTopicsList []string

// AddIfNotNull appends the topic to the list if the function is not nil
func (e *ExtraTopicsList) AddIfNotNull(fn interface{}, topic string) {
	if fn != nil {
		*e = append(*e, topic)
	}
}

type ClimateMeta struct {
	Meta

	MinTemp   decimal.Decimal
	MaxTemp   decimal.Decimal
	Increment decimal.Decimal
	Precision decimal.Decimal

	Modes       []string
	PresetModes []string

	ExtraTopics ExtraTopicsList
}

func (c ClimateMeta) GetId() string {
	return c.Id
}

func (c ClimateMeta) GetName() string {
	return c.Name
}

func (c ClimateMeta) GetExtraTopics() []string {
	return c.ExtraTopics
}

func (c ClimateMeta) ToHaDiscovery(dev Device) map[string]any {
	result := c.Meta.ToHaDiscovery(dev, "climate")

	if !c.MinTemp.IsZero() {
		result["min_temp"] = c.MinTemp.String()
	}
	if !c.MaxTemp.IsZero() {
		result["max_temp"] = c.MaxTemp.String()
	}
	if !c.Increment.IsZero() {
		result["temp_step"] = c.Increment.InexactFloat64()
	}
	if !c.Precision.IsZero() {
		result["precision"] = c.Precision.InexactFloat64()
	}

	if len(c.Modes) > 0 {
		result["modes"] = c.Modes
	}
	if len(c.PresetModes) > 0 {
		result["pr_modes"] = c.PresetModes
	}

	return result
}

type Climate struct {
	Meta                 observable.Observable[ClimateMeta]
	ActualTemperature    observable.Observable[decimal.Decimal]
	RequestedTemperature observable.Observable[decimal.Decimal]
	Mode                 observable.Observable[string]
	PresetMode           observable.Observable[string]
	Attributes           observable.Observable[Attrs]

	SetTemperature func(decimal.Decimal) error
	SetMode        func(string) error
	SetPresetMode  func(string) error
}
