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

// TestStandardNumberEntity_PublishesInitialState verifies that standard
// number entities publish their initial state immediately.
func TestStandardNumberEntity_PublishesInitialState(t *testing.T) {
	nc, ns := startTestNATS(t)
	defer ns.Shutdown()
	reg := newTestRegistry(t, nc)

	stateTopic := executor.Topic(testOriginName, testDeviceName, "number", "temperature", "state")

	// Create and add number entity
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
				Id:   "test_sensor",
				Name: "Test Sensor",
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
