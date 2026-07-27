package tests

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// StartTestNATS starts an embedded NATS server for testing.
// Call t.Cleanup to shut it down.
func StartTestNATS(t *testing.T) (*nats.Conn, *server.Server) {
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

// WaitForRetained waits up to 3 seconds for a message to be delivered to the given topic.
// Returns the message data or an error if no message was received.
func WaitForRetained(t *testing.T, nc *nats.Conn, topic string) string {
	t.Helper()
	sub, err := nc.SubscribeSync(topic)
	if err != nil {
		t.Fatalf("failed to subscribe to %s: %v", topic, err)
	}
	defer sub.Unsubscribe()

	msg, err := sub.NextMsg(3 * time.Second)
	if err != nil {
		t.Fatalf("timed out waiting for message on %s: %v", topic, err)
	}
	return string(msg.Data)
}

// PublishRetained publishes a retained message to the given topic.
func PublishRetained(t *testing.T, nc *nats.Conn, topic, value string) {
	t.Helper()
	if err := nc.Publish(topic, []byte(value)); err != nil {
		t.Fatalf("failed to publish retained to %s: %v", topic, err)
	}
}

// RequestRetained publishes a retained value request and waits for the bridge to respond.
func RequestRetained(t *testing.T, nc *nats.Conn, topic string) {
	t.Helper()
	retainedTopic := "mqtt-retained." + topic
	if err := nc.Publish(retainedTopic, []byte{}); err != nil {
		t.Fatalf("failed to request retained for %s: %v", topic, err)
	}
}

// ContextWithTimeout returns a context with a 2-second timeout, appropriate for short test operations.
func ContextWithTimeout(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 2*time.Second)
}

// ContextWithCancel returns a cancellable context. Call t.Cleanup(cancel) to release resources.
func ContextWithCancel(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithCancel(context.Background())
}
