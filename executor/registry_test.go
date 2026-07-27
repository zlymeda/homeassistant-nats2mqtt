package executor

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

const (
	testDeviceID = "test-vehicle-001"
	originName   = "test-origin"
	deviceName   = "test-vehicle"
)

func startTestNATS(t *testing.T) (*nats.Conn, *server.Server) {
	t.Helper()
	opts := &server.Options{
		Host: "127.0.0.1",
		Port: -1,
	}
	ns, err := server.NewServer(opts)
	if err != nil {
		t.Fatalf("failed to create nats server: %v", err)
	}
	ns.Start()
	if !ns.ReadyForConnections(5 * time.Second) {
		t.Fatal("nats server not ready")
	}
	t.Cleanup(func() { ns.Shutdown() })

	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	t.Cleanup(func() { nc.Close() })
	return nc, ns
}

func newTestRegistry(t *testing.T, nc *nats.Conn) *EntityRegistry {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	return &EntityRegistry{
		ctx:          ctx,
		nc:           nc,
		device:       entity.Device{Id: testDeviceID, Name: deviceName},
		topicPrefix:  Topic(originName, deviceName),
		origin:       Origin{Name: originName},
		stateUpdated: func() {},
	}
}

// TestMonitorMqttOnlyState_RetainedValuePublishesInitialValue verifies that when a retained value exists,
// it's received and published without publishing the default.
func TestMonitorMqttOnlyState_RetainedValuePublishesInitialValue(t *testing.T) {
	nc, _ := startTestNATS(t)
	reg := newTestRegistry(t, nc)

	rawMeta := observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
		Meta: entity.Meta{
			Id:             "solar_power",
			Name:           "Solar Power",
			EntityCategory: entity.CategoryConfig,
			MqttOnly:       true,
		},
	})
	meta := observable.NewMapped(rawMeta, func(m entity.NumberMeta) entity.Metadata { return m })
	state := observable.NewSimple(context.Background(), "0.0")

	// Subscribe to retained values
	stateTopic := "test-origin.test-vehicle.number.solar_power.state"
	retainedTopic := "mqtt-retained." + stateTopic

	// Subscribe to retained requests
	_, err := nc.Subscribe(retainedTopic, func(msg *nats.Msg) {
		nc.Publish(stateTopic, []byte("100.0"))
	})
	if err != nil {
		t.Fatalf("failed to subscribe to retained topic: %v", err)
	}

	// Start monitoring
	reg.monitorMqttOnlyState(meta, state, "")

	// Wait for retained value to arrive
	time.Sleep(500 * time.Millisecond)

	if got := state.Current(); got != "100.0" {
		t.Errorf("expected state 100.0, got %s", got)
	}
}

// TestMonitorMqttOnlyState_NoRetainedValue verifies that when no retained value exists,
// the default value is published after the timeout.
func TestMonitorMqttOnlyState_NoRetainedValue(t *testing.T) {
	nc, _ := startTestNATS(t)
	reg := newTestRegistry(t, nc)
	rawMeta := observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
		Meta: entity.Meta{
			Id:       "charge_mode",
			Name:     "Charging Mode",
			MqttOnly: true,
		},
	})
	meta := observable.NewMapped(rawMeta, func(m entity.NumberMeta) entity.Metadata { return m })
	state := observable.NewSimple(context.Background(), "normal")

	// No retained value request - the default should be published after timeout
	stateTopic := "test-origin.test-vehicle.number.charge_mode.state"
	retainedTopic := "mqtt-retained." + stateTopic

	// Subscribe to retained request
	_, err := nc.Subscribe(retainedTopic, func(msg *nats.Msg) {
		nc.Publish(stateTopic, []byte("night"))
	})
	if err != nil {
		t.Fatalf("failed to subscribe to retained topic: %v", err)
	}

	// Start monitoring
	reg.monitorMqttOnlyState(meta, state, "")

	// Wait for default value to be published (2s timeout + 100ms buffer)
	time.Sleep(2500 * time.Millisecond)

	if got := state.Current(); got != "night" {
		t.Errorf("expected state 'night', got %s", got)
	}
}

// TestMonitorMqttOnlyState_ImmediateRetainedValue verifies that retained values are received immediately
func TestMonitorMqttOnlyState_ImmediateRetainedValue(t *testing.T) {
	nc, _ := startTestNATS(t)
	reg := newTestRegistry(t, nc)

	rawMeta := observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
		Meta: entity.Meta{
			Id:       "charging_mode",
			Name:     "Charging Mode",
			MqttOnly: true,
		},
	})
	meta := observable.NewMapped(rawMeta, func(m entity.NumberMeta) entity.Metadata { return m })
	state := observable.NewSimple(context.Background(), "normal")

	// Subscribe to retained request - respond with a value
	stateTopic := "test-origin.test-vehicle.number.charging_mode.state"
	retainedTopic := "mqtt-retained." + stateTopic

	_, err := nc.Subscribe(retainedTopic, func(msg *nats.Msg) {
		nc.Publish(stateTopic, []byte("charging"))
	})
	if err != nil {
		t.Fatalf("failed to subscribe to retained topic: %v", err)
	}

	// Start monitoring - triggers retained request
	reg.monitorMqttOnlyState(meta, state, "")

	// Wait for retained value to arrive
	time.Sleep(500 * time.Millisecond)

	if got := state.Current(); got != "charging" {
		t.Errorf("expected state 'charging', got %s", got)
	}
}

// TestMonitorObservable_PublishesInitialState verifies that initial state is published when monitorObservable is called
func TestMonitorObservable_PublishesInitialState(t *testing.T) {
	nc, _ := startTestNATS(t)
	reg := newTestRegistry(t, nc)
	rawMeta := observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
		Meta: entity.Meta{
			Id:   "temperature",
			Name: "Temperature",
		},
	})
	meta := observable.NewMapped(rawMeta, func(m entity.NumberMeta) entity.Metadata { return m })
	state := observable.NewSimple(context.Background(), "22.5")

	stateTopic := "test-origin.test-vehicle.number.temperature.state"

	monitorObservable(reg, meta, state, func(state string) error {
		return nc.Publish(stateTopic, []byte(state))
	})

	reg.PublishStates()
	time.Sleep(200 * time.Millisecond)
}
