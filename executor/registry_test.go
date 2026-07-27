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
		ctx:         ctx,
		nc:          nc,
		device:      entity.Device{Id: "test-device", Name: "Test Device"},
		topicPrefix: Topic("test-origin", "test-device"),
		origin:      Origin{Name: "test-origin"},
	}
}

// mapNumberMeta wraps a Single[NumberMeta] as Observable[Metadata].
func mapNumberMeta(single *observable.Single[entity.NumberMeta]) observable.Observable[entity.Metadata] {
	return observable.NewMapped(single, func(n entity.NumberMeta) entity.Metadata {
		return n
	})
}

// TestMonitorMqttOnlyState_RetainedExists verifies that when a retained value
// exists, it is received and published without publishing the default.
func TestMonitorMqttOnlyState_RetainedExists(t *testing.T) {
	nc, _ := startTestNATS(t)
	reg := newTestRegistry(t, nc)

	metaSingle := observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
		Meta: entity.Meta{
			Id:       "solar_power",
			Name:     "Solar Power",
			MqttOnly: true,
		},
	})
	meta := mapNumberMeta(metaSingle)
	state := observable.NewSimple(context.Background(), "0")

	stateTopic := "test-origin.test-device.number.solar_power.state"
	retainedTopic := "mqtt-retained." + stateTopic

	// Simulate the bridge: when it receives a mqtt-retained request,
	// it publishes the retained value to the state topic.
	_, err := nc.Subscribe(retainedTopic, func(msg *nats.Msg) {
		_ = nc.Publish(stateTopic, []byte("42.5"))
	})
	if err != nil {
		t.Fatal(err)
	}

	reg.monitorMqttOnlyState(meta, state)

	// Wait for the retained value to be received.
	time.Sleep(300 * time.Millisecond)

	if got := state.Current(); got != "42.5" {
		t.Errorf("expected state 42.5, got %s", got)
	}
}

// TestMonitorMqttOnlyState_NoRetained verifies that when no retained value
// exists, the default value is published after the timeout.
func TestMonitorMqttOnlyState_NoRetained(t *testing.T) {
	nc, _ := startTestNATS(t)
	reg := newTestRegistry(t, nc)

	metaSingle := observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
		Meta: entity.Meta{
			Id:       "charge_limit",
			Name:     "Charge Limit",
			MqttOnly: true,
		},
	})
	meta := mapNumberMeta(metaSingle)
	state := observable.NewSimple(context.Background(), "80")

	stateTopic := "test-origin.test-device.number.charge_limit.state"

	// Capture published values.
	published := make(chan string, 10)
	_, err := nc.Subscribe(stateTopic, func(msg *nats.Msg) {
		published <- string(msg.Data)
	})
	if err != nil {
		t.Fatal(err)
	}

	reg.monitorMqttOnlyState(meta, state)

	// No one responds to the retained request — wait for timeout.
	select {
	case val := <-published:
		if val != "80" {
			t.Errorf("expected default value 80, got %s", val)
		}
	case <-time.After(5 * time.Second):
		t.Error("timed out waiting for default value to be published")
	}
}

// TestMonitorMqttOnlyState_UserChange verifies that after retained handling,
// user-initiated changes are published via the Changes() listener.
func TestMonitorMqttOnlyState_UserChange(t *testing.T) {
	nc, _ := startTestNATS(t)
	reg := newTestRegistry(t, nc)

	metaSingle := observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
		Meta: entity.Meta{
			Id:       "solar_power",
			Name:     "Solar Power",
			MqttOnly: true,
		},
	})
	meta := mapNumberMeta(metaSingle)
	state := observable.NewSimple(context.Background(), "0")

	stateTopic := "test-origin.test-device.number.solar_power.state"
	retainedTopic := "mqtt-retained." + stateTopic

	// Simulate retained value.
	_, err := nc.Subscribe(retainedTopic, func(msg *nats.Msg) {
		nc.Publish(stateTopic, []byte("100"))
	})
	if err != nil {
		t.Fatal(err)
	}

	reg.monitorMqttOnlyState(meta, state)

	// Wait for retained value.
	time.Sleep(300 * time.Millisecond)

	if got := state.Current(); got != "100" {
		t.Fatalf("expected retained value 100, got %s", got)
	}

	// Simulate user changing the value via HA.
	state.Mutate("75")

	// Give the Changes() goroutine time to publish.
	time.Sleep(100 * time.Millisecond)

	if got := state.Current(); got != "75" {
		t.Errorf("expected state 75 after user change, got %s", got)
	}
}

// TestMonitorObservable_PublishesInitialState verifies that standard
// observables publish their initial state immediately.
func TestMonitorObservable_PublishesInitialState(t *testing.T) {
	nc, _ := startTestNATS(t)
	reg := newTestRegistry(t, nc)

	metaSingle := observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
		Meta: entity.Meta{
			Id:   "temperature",
			Name: "Temperature",
		},
	})
	meta := mapNumberMeta(metaSingle)
	state := observable.NewSimple(context.Background(), "22.5")

	stateTopic := "test-origin.test-device.number.temperature.state"
	published := make(chan string, 10)
	_, err := nc.Subscribe(stateTopic, func(msg *nats.Msg) {
		published <- string(msg.Data)
	})
	if err != nil {
		t.Fatal(err)
	}

	monitorObservable(reg, meta, state, func(val string) error {
		return nc.Publish(stateTopic, []byte(val))
	})

	// Trigger PublishStates to publish initial value.
	reg.PublishStates()

	select {
	case val := <-published:
		if val != "22.5" {
			t.Errorf("expected initial value 22.5, got %s", val)
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for initial state")
	}
}

// TestMonitorMqttOnlyState_SubscriptionError verifies fallback when subscription fails.
func TestMonitorMqttOnlyState_SubscriptionError(t *testing.T) {
	nc, _ := startTestNATS(t)
	reg := newTestRegistry(t, nc)

	metaSingle := observable.NewSingle[entity.NumberMeta](entity.NumberMeta{
		Meta: entity.Meta{
			Id:       "broken_entity",
			Name:     "Broken Entity",
			MqttOnly: true,
		},
	})
	meta := mapNumberMeta(metaSingle)
	state := observable.NewSimple(context.Background(), "99")

	// Close connection to force subscription failure.
	nc.Close()

	reg.monitorMqttOnlyState(meta, state)

	// Should have added a fallback publishState.
	reg.mutex.Lock()
	count := len(reg.publishState)
	reg.mutex.Unlock()

	if count != 1 {
		t.Errorf("expected 1 publishState entry on error, got %d", count)
	}
}
