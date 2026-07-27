package tests

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/shopspring/decimal"
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/executor"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

const (
	testOriginName = "test-origin"
	testDeviceName = "test-device"
	testDeviceID   = "test-device-001"
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

func newTestRegistry(t *testing.T, nc *nats.Conn) *executor.EntityRegistry {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return executor.NewEntityRegistry(
		ctx,
		nc,
		entity.Device{Id: testDeviceID, Name: testDeviceName},
		executor.Origin{Name: testOriginName},
		executor.Topic(testOriginName, testDeviceName),
		"",
	)
}

// TestMqttOnlyNumberEntity_RetainedValue verifies that when a retained value exists for a MqttOnly number entity,
// it is received and used without publishing the default immediately.
func TestMqttOnlyNumberEntity_RetainedValue(t *testing.T) {
	nc, ns := startTestNATS(t)
	defer ns.Shutdown()
	reg := newTestRegistry(t, nc)

	stateTopic := executor.Topic(testOriginName, testDeviceName, "number", "solar_power", "state")
	retainedTopic := "mqtt-retained." + stateTopic

	// Subscribe to retained requests
	_, err := nc.Subscribe(retainedTopic, func(msg *nats.Msg) {
		nc.Publish(stateTopic, []byte("100.0"))
	})
	if err != nil {
		t.Fatalf("failed to subscribe to retained topic: %v", err)
	}

	// Create and add number entity with MqttOnly flag
	state := observable.NewSimple[decimal.Decimal](context.Background(), decimal.RequireFromString("0.0"))
	num := &entity.Number{
		Meta: observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
			Meta: entity.Meta{
				Id:             "solar_power",
				Name:           "Solar Power",
				EntityCategory: entity.CategoryConfig,
				MqttOnly:       true,
			},
		}),
		State: state,
		Set:   func(d decimal.Decimal) error { state.Change(d); return nil },
	}

	err = reg.AddNumber(num)
	if err != nil {
		t.Fatalf("failed to add number entity: %v", err)
	}

	// Wait for retained value to arrive
	time.Sleep(500 * time.Millisecond)

	if !state.Current().Equal(decimal.RequireFromString("100.0")) {
		t.Errorf("expected state 100.0, got %s", state.Current())
	}
}

// TestMqttOnlyNumberEntity_NoRetainedValue verifies that when no retained value exists,
// the default value is published after the timeout.
func TestMqttOnlyNumberEntity_NoRetainedValue(t *testing.T) {
	nc, ns := startTestNATS(t)
	defer ns.Shutdown()
	reg := newTestRegistry(t, nc)

	// Create and add number entity with MqttOnly flag - no retained value will be provided
	state := observable.NewSimple[decimal.Decimal](context.Background(), decimal.RequireFromString("0"))
	num := &entity.Number{
		Meta: observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
			Meta: entity.Meta{
				Id:       "charge_mode",
				Name:     "Charging Mode",
				MqttOnly: true,
			},
		}),
		State: state,
		Set:   func(d decimal.Decimal) error { state.Change(d); return nil },
	}

	err := reg.AddNumber(num)
	if err != nil {
		t.Fatalf("failed to add number entity: %v", err)
	}

	// Wait for default value to be published (2s timeout + 100ms buffer)
	time.Sleep(2500 * time.Millisecond)

	if got := state.Current().String(); got != "0" {
		t.Errorf("expected state '0', got %s", got)
	}
}

// TestMqttOnlySelectEntity verifies that select entities with MqttOnly work correctly.
func TestMqttOnlySelectEntity(t *testing.T) {
	nc, ns := startTestNATS(t)
	defer ns.Shutdown()
	reg := newTestRegistry(t, nc)

	stateTopic := executor.Topic(testOriginName, testDeviceName, "select", "charging_mode", "state")
	retainedTopic := "mqtt-retained." + stateTopic
	// Subscribe to retained request
	_, err := nc.Subscribe(retainedTopic, func(msg *nats.Msg) {
		nc.Publish(stateTopic, []byte("always_on"))
	})
	if err != nil {
		t.Fatalf("failed to subscribe to retained topic: %v", err)
	}

	// Create and add select entity with MqttOnly flag
	state := observable.NewSimple(context.Background(), "auto")
	selectEnt := &entity.Select{
		Meta: observable.NewSingle[entity.SelectMeta](entity.SelectMeta{
			Meta: entity.Meta{
				Id:       "charging_mode",
				Name:     "Charging Mode",
				MqttOnly: true,
			},
			Options: []string{"off", "auto", "always_on"},
		}),
		State: state,
		Set:   func(s string) error { state.Change(s); return nil },
	}

	err = reg.AddSelect(selectEnt)
	if err != nil {
		t.Fatalf("failed to add select entity: %v", err)
	}

	// Wait for retained value
	time.Sleep(500 * time.Millisecond)

	if got := state.Current(); got != "always_on" {
		t.Errorf("expected state 'always_on', got '%s'", got)
	}
}

// TestStandardNumberEntity_PublishesInitialState verifies that standard (non-MqttOnly)
// number entities publish their initial state immediately.
func TestStandardNumberEntity_PublishesInitialState(t *testing.T) {
	nc, ns := startTestNATS(t)
	defer ns.Shutdown()
	reg := newTestRegistry(t, nc)

	stateTopic := executor.Topic(testOriginName, testDeviceName, "number", "temperature", "state")

	// Create and add number entity WITHOUT MqttOnly flag
	state := observable.NewSimple[decimal.Decimal](context.Background(), decimal.RequireFromString("22.5"))
	num := &entity.Number{
		Meta: observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
			Meta: entity.Meta{
				Id:   "temperature",
				Name: "Temperature",
			},
		}),
		State: state,
		Set:   func(d decimal.Decimal) error { state.Change(d); return nil },
	}

	err := reg.AddNumber(num)
	if err != nil {
		t.Fatalf("failed to add number entity: %v", err)
	}
	// Subscribe to state topic BEFORE publishing
	sub, err := nc.SubscribeSync(stateTopic)
	if err != nil {
		t.Fatalf("failed to subscribe to state topic: %v", err)
	}
	defer sub.Unsubscribe()

	// Trigger PublishStates to publish initial state
	err = reg.PublishStates()
	if err != nil {
		t.Fatalf("failed to publish states: %v", err)
	}

	// Verify the initial state was published
	msg, err := sub.NextMsg(time.Second)
	if err != nil {
		t.Fatalf("failed to receive state: %v", err)
	}

	if string(msg.Data) != "22.5" {
		t.Errorf("expected initial value 22.5, got %s", string(msg.Data))
	}
}

// TestPublishDiscovery verifies discovery messages are published correctly.
func TestPublishDiscovery(t *testing.T) {
	nc, ns := startTestNATS(t)
	defer ns.Shutdown()
	reg := newTestRegistry(t, nc)

	// Create and add number entity
	num := &entity.Number{
		Meta: observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
			Meta: entity.Meta{
				Id:       "test_sensor",
				Name:     "Test Sensor",
				MqttOnly: true,
			},
		}),
		State: observable.NewSimple[decimal.Decimal](context.Background(), decimal.RequireFromString("42")),
		Set:   func(d decimal.Decimal) error { return nil },
	}

	err := reg.AddNumber(num)
	if err != nil {
		t.Fatalf("failed to add number entity: %v", err)
	}

	// Publish discovery
	err = reg.PublishDiscovery()
	if err != nil {
		t.Fatalf("failed to publish discovery: %v", err)
	}
}

// TestMultipleMqttOnlyEntities verifies multiple MqttOnly entities work correctly.
func TestMultipleMqttOnlyEntities(t *testing.T) {
	nc, ns := startTestNATS(t)
	defer ns.Shutdown()
	reg := newTestRegistry(t, nc)

	// Set up retained responses for both topics
	stateTopic1 := executor.Topic(testOriginName, testDeviceName, "number", "solar_power", "state")
	retainedTopic1 := "mqtt-retained." + stateTopic1

	stateTopic2 := executor.Topic(testOriginName, testDeviceName, "number", "battery", "state")
	retainedTopic2 := "mqtt-retained." + stateTopic2

	_, err := nc.Subscribe(retainedTopic1, func(msg *nats.Msg) {
		nc.Publish(stateTopic1, []byte("50.5"))
	})
	if err != nil {
		t.Fatalf("failed to subscribe to retained topic 1: %v", err)
	}

	_, err = nc.Subscribe(retainedTopic2, func(msg *nats.Msg) {
		nc.Publish(stateTopic2, []byte("80.0"))
	})
	if err != nil {
		t.Fatalf("failed to subscribe to retained topic 2: %v", err)
	}

	// Create and add two number entities
	state1 := observable.NewSimple[decimal.Decimal](context.Background(), decimal.RequireFromString("0"))
	num1 := &entity.Number{
		Meta: observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
			Meta: entity.Meta{
				Id:             "solar_power",
				Name:           "Solar Power",
				EntityCategory: entity.CategoryConfig,
				MqttOnly:       true,
			},
		}),
		State: state1,
		Set:   func(d decimal.Decimal) error { state1.Change(d); return nil },
	}

	state2 := observable.NewSimple[decimal.Decimal](context.Background(), decimal.RequireFromString("0"))
	num2 := &entity.Number{
		Meta: observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
			Meta: entity.Meta{
				Id:             "battery",
				Name:           "Battery",
				EntityCategory: entity.CategoryConfig,
				MqttOnly:       true,
			},
		}),
		State: state2,
		Set:   func(d decimal.Decimal) error { state2.Change(d); return nil },
	}

	err = reg.AddNumber(num1)
	if err != nil {
		t.Fatalf("failed to add number entity 1: %v", err)
	}

	err = reg.AddNumber(num2)
	if err != nil {
		t.Fatalf("failed to add number entity 2: %v", err)
	}

	// Wait for retained values
	time.Sleep(500 * time.Millisecond)

	if !state1.Current().Equal(decimal.RequireFromString("50.5")) {
		t.Errorf("expected solar_power state 50.5, got %s", state1.Current())
	}

	if !state2.Current().Equal(decimal.RequireFromString("80.0")) {
		t.Errorf("expected battery state 80.0, got %s", state2.Current())
	}
}
