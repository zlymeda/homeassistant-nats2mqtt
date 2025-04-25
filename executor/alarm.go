package executor

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddAlarm(e *entity.Alarm) error {
	alarmMeta := observable.NewMapped(e.Meta, func(i entity.Meta) entity.Metadata {
		meta := entity.AlarmMeta{
			Meta: i,
		}

		meta.SupportedFeatures.AddIfNotNull(e.ArmAway, entity.AlarmFeatureArmAway)
		meta.SupportedFeatures.AddIfNotNull(e.ArmHome, entity.AlarmFeatureArmHome)
		meta.SupportedFeatures.AddIfNotNull(e.ArmNight, entity.AlarmFeatureArmNight)
		meta.SupportedFeatures.AddIfNotNull(e.ArmVacation, entity.AlarmFeatureArmVacation)
		meta.SupportedFeatures.AddIfNotNull(e.ArmCustomBypass, entity.AlarmFeatureArmCustomBypass)

		return meta
	})

	s.register(entity.Entity{
		Meta: alarmMeta,
		State: observable.NewMapped(e.State, func(i entity.AlarmState) string {
			return string(i)
		}),
		Attributes: e.Attributes,
	})

	commands := map[string]func() error{
		"ARM_AWAY":          e.ArmAway,
		"ARM_HOME":          e.ArmHome,
		"ARM_NIGHT":         e.ArmNight,
		"ARM_VACATION":      e.ArmVacation,
		"ARM_CUSTOM_BYPASS": e.ArmCustomBypass,
		"DISARM":            e.Disarm,
	}

	meta := alarmMeta.Current()

	cmdTopic := s.fullTopic(meta, entity.CommandTopic)
	if err := s.monitorCommandsOn(cmdTopic, createCallback(commands)); err != nil {
		return err
	}

	return nil
}
