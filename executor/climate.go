package executor

import (
	"errors"
	"github.com/shopspring/decimal"
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddClimate(e *entity.Climate) error {
	s.registerClimate(e)

	meta := e.Meta.Current()

	s.monitorDecimal(meta, e.ActualTemperature, entity.ClimateCurrentTemperatureTopic)
	s.monitorDecimal(meta, e.RequestedTemperature, entity.ClimateTemperatureStateTopic)
	s.monitorString(meta, e.Mode, entity.ClimateModeStateTopic)
	s.monitorString(meta, e.PresetMode, entity.ClimatePresetModeStateTopic)

	err1 := s.monitorDecimalCmd(meta, e.SetTemperature, entity.ClimateTemperatureCommandTopic)
	err2 := s.monitorStringCmd(meta, e.SetMode, entity.ClimateModeCommandTopic)
	err3 := s.monitorStringCmd(meta, e.SetPresetMode, entity.ClimatePresetModeCommandTopic)

	return errors.Join(err1, err2, err3)
}

func (s *EntityRegistry) registerClimate(e *entity.Climate) {
	s.register(entity.Entity{
		Meta: observable.NewMapped(e.Meta, func(i entity.ClimateMeta) entity.Metadata {

			i.ExtraTopics.AddIfNotNull(e.ActualTemperature, entity.ClimateCurrentTemperatureTopic)
			i.ExtraTopics.AddIfNotNull(e.RequestedTemperature, entity.ClimateTemperatureStateTopic)
			i.ExtraTopics.AddIfNotNull(e.Mode, entity.ClimateModeStateTopic)
			i.ExtraTopics.AddIfNotNull(e.PresetMode, entity.ClimatePresetModeStateTopic)
			i.ExtraTopics.AddIfNotNull(e.SetTemperature, entity.ClimateTemperatureCommandTopic)
			i.ExtraTopics.AddIfNotNull(e.SetMode, entity.ClimateModeCommandTopic)
			i.ExtraTopics.AddIfNotNull(e.SetPresetMode, entity.ClimatePresetModeCommandTopic)

			return i
		}),
		Attributes: e.Attributes,
	})
}

func (s *EntityRegistry) monitorDecimal(meta entity.Metadata, temperature observable.Observable[decimal.Decimal], subTopic string) {
	if temperature == nil {
		return
	}

	topic := s.fullTopic(meta, subTopic)

	monitorObservable(s, meta, temperature, func(temp decimal.Decimal) error {
		return s.nc.Publish(topic, []byte(temp.String()))
	})
}

func (s *EntityRegistry) monitorString(meta entity.Metadata, str observable.Observable[string], subTopic string) {
	if str == nil {
		return
	}

	topic := s.fullTopic(meta, subTopic)

	monitorObservable(s, meta, str, func(value string) error {
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

func (s *EntityRegistry) monitorStringCmd(meta entity.ClimateMeta, set func(string) error, subTopic string) error {
	if set == nil {
		return nil
	}

	topic := s.fullTopic(meta, subTopic)

	return s.monitorCommandsOn(topic, func(payload []byte) error {
		return set(string(payload))
	})
}
