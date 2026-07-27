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
