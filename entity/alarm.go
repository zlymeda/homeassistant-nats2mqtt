package entity

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

var _ Metadata = AlarmMeta{}

// SupportedFeaturesList is a helper type that extends []string with an AddIfNotNull method
type SupportedFeaturesList []string

// AddIfNotNull appends the feature to the list if the function is not nil
func (s *SupportedFeaturesList) AddIfNotNull(fn func() error, feature string) {
	if fn != nil {
		*s = append(*s, feature)
	}
}

type AlarmMeta struct {
	Meta

	SupportedFeatures SupportedFeaturesList
}

func (a AlarmMeta) GetId() string {
	return a.Id
}

func (a AlarmMeta) GetName() string {
	return a.Name
}

func (a AlarmMeta) GetExtraTopics() []string {
	return []string{CommandTopic}
}

func (a AlarmMeta) ToHaDiscovery(dev Device) map[string]any {
	result := a.Meta.ToHaDiscovery(dev, "alarm_control_panel")

	// Add alarm-specific fields
	result["cod_arm_req"] = false
	result["cod_dis_req"] = false
	result["cod_trig_req"] = false

	if len(a.SupportedFeatures) > 0 {
		result["supported_features"] = a.SupportedFeatures
	}

	return result
}

type AlarmState string
type AlarmFeature = string

const (
	AlarmFeatureArmAway         AlarmFeature = "arm_away"
	AlarmFeatureArmHome         AlarmFeature = "arm_home"
	AlarmFeatureArmNight        AlarmFeature = "arm_night"
	AlarmFeatureArmVacation     AlarmFeature = "arm_vacation"
	AlarmFeatureArmCustomBypass AlarmFeature = "arm_custom_bypass"

	AlarmStateDisarmed          AlarmState = "disarmed"
	AlarmStateArmedAway         AlarmState = "armed_away"
	AlarmStateArmedHome         AlarmState = "armed_home"
	AlarmStateArmedNight        AlarmState = "armed_night"
	AlarmStateArmedVacation     AlarmState = "armed_vacation"
	AlarmStateArmedCustomBypass AlarmState = "armed_custom_bypass"
	AlarmStateTriggered         AlarmState = "triggered"
)

type Alarm struct {
	Meta       observable.Observable[Meta]
	State      observable.Observable[AlarmState]
	Attributes observable.Observable[Attrs]

	ArmAway         func() error
	ArmHome         func() error
	ArmNight        func() error
	ArmVacation     func() error
	ArmCustomBypass func() error
	Disarm          func() error
}
