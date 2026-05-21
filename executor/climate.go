package executor

import (
	"errors"

	"github.com/shopspring/decimal"
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddClimate(e *entity.Climate) error {
	climateMeta := s.registerClimate(e)

	meta := climateMeta.Current()

	s.monitorDecimal(climateMeta, e.ActualTemperature, entity.ClimateCurrentTemperatureTopic)
	s.monitorDecimal(climateMeta, e.RequestedTemperature, entity.ClimateTemperatureStateTopic)
	s.monitorString(climateMeta, e.Mode, entity.ClimateModeStateTopic)
	s.monitorString(climateMeta, e.PresetMode, entity.ClimatePresetModeStateTopic)

	err1 := s.monitorDecimalCmd(meta, e.SetTemperature, entity.ClimateTemperatureCommandTopic)
	err2 := s.monitorStringCmd(meta, e.SetMode, entity.ClimateModeCommandTopic)
	err3 := s.monitorStringCmd(meta, e.SetPresetMode, entity.ClimatePresetModeCommandTopic)

	return errors.Join(err1, err2, err3)
}

func (s *EntityRegistry) registerClimate(e *entity.Climate) observable.Observable[entity.Metadata] {
	climateMeta := observable.NewMapped(e.Meta, func(i entity.ClimateMeta) entity.Metadata {

		i.ExtraTopics.AddIfNotNull(e.ActualTemperature, entity.ClimateCurrentTemperatureTopic)
		i.ExtraTopics.AddIfNotNull(e.RequestedTemperature, entity.ClimateTemperatureStateTopic)
		i.ExtraTopics.AddIfNotNull(e.Mode, entity.ClimateModeStateTopic)
		i.ExtraTopics.AddIfNotNull(e.PresetMode, entity.ClimatePresetModeStateTopic)
		i.ExtraTopics.AddIfNotNull(e.SetTemperature, entity.ClimateTemperatureCommandTopic)
		i.ExtraTopics.AddIfNotNull(e.SetMode, entity.ClimateModeCommandTopic)
		i.ExtraTopics.AddIfNotNull(e.SetPresetMode, entity.ClimatePresetModeCommandTopic)

		return i
	})

	s.register(entity.Entity{
		Meta:       climateMeta,
		Attributes: e.Attributes,
	})

	return climateMeta
}

func (s *EntityRegistry) monitorDecimal(meta observable.Observable[entity.Metadata], temperature observable.Observable[decimal.Decimal], subTopic string) {
	if temperature == nil {
		return
	}

	monitorObservable(s, meta, temperature, func(temp decimal.Decimal) error {
		currentMeta := meta.Current()
		topic := s.fullTopic(currentMeta, subTopic)
		return s.nc.Publish(topic, []byte(temp.String()))
	})
}

func (s *EntityRegistry) monitorString(meta observable.Observable[entity.Metadata], str observable.Observable[string], subTopic string) {
	if str == nil {
		return
	}

	monitorObservable(s, meta, str, func(value string) error {
		currentMeta := meta.Current()
		topic := s.fullTopic(currentMeta, subTopic)
		return s.nc.Publish(topic, []byte(value))
	})
}

func (s *EntityRegistry) monitorDecimalCmd(meta entity.Metadata, set func(decimal.Decimal) error, subTopic string) error {
	if set == nil {
		return nil
	}

	topic := s.fullTopic(meta, subTopic)

	return s.monitorCommandsOn(topic, func(payload []byte) error {
		temp, err := decimal.NewFromString(string(payload))
		if err != nil {
			return err
		}
		return set(temp)
	})
}

func (s *EntityRegistry) monitorStringCmd(meta entity.Metadata, set func(string) error, subTopic string) error {
	if set == nil {
		return nil
	}

	topic := s.fullTopic(meta, subTopic)

	return s.monitorCommandsOn(topic, func(payload []byte) error {
		return set(string(payload))
	})
}
