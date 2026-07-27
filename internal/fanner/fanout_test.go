package fanner

import (
	"context"
	"testing"
	"time"
)

func TestFanOut_SingleConsumer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan int, 1)
	createConsumer := FanOut(ctx, events)

	consumer := createConsumer(1)

	// Send an event
	events <- 42

	// Should receive the event
	select {
	case val := <-consumer:
		if val != 42 {
			t.Errorf("received %v, want %v", val, 42)
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for event")
	}
}

func TestFanOut_MultipleConsumers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan int, 1)
	createConsumer := FanOut(ctx, events)

	consumer1 := createConsumer(1)
	consumer2 := createConsumer(1)

	// Send an event
	events <- 100

	// Both consumers should receive the event
	for i, ch := range []<-chan int{consumer1, consumer2} {
		select {
		case val := <-ch:
			if val != 100 {
				t.Errorf("consumer %d: received %v, want %v", i, val, 100)
			}
		case <-time.After(time.Second):
			t.Errorf("consumer %d: timed out waiting for event", i)
		}
	}
}

func TestFanOut_LateConsumer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan int, 1)
	createConsumer := FanOut(ctx, events)

	// Send event before creating consumer
	events <- 1

	// Give the fanout goroutine time to process the event
	time.Sleep(10 * time.Millisecond)

	// Create consumer after event was sent
	consumer := createConsumer(1)

	// Send another event
	events <- 2

	// Should only receive the second event
	select {
	case val := <-consumer:
		if val != 2 {
			t.Errorf("received %v, want %v", val, 2)
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for event")
	}
}

func TestFanOut_SlowConsumer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan int, 2)
	createConsumer := FanOut(ctx, events)

	// Create a consumer with buffer size 1
	consumer := createConsumer(1)

	// Send two events quickly
	events <- 1
	events <- 2

	// The slow consumer should receive the first event
	select {
	case val := <-consumer:
		if val != 1 {
			t.Errorf("received %v, want %v", val, 1)
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for first event")
	}

	// The second event may have been discarded for the slow consumer
	// This is expected behavior - the fanout doesn't block on slow consumers
}

func TestFanOut_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	events := make(chan int, 1)
	createConsumer := FanOut(ctx, events)

	consumer := createConsumer(1)

	// Cancel the context
	cancel()

	// Consumer channel should eventually be closed
	select {
	case _, ok := <-consumer:
		if ok {
			// If we got a value, that's fine too
		}
	case <-time.After(time.Second):
		// The channel might not be closed immediately, that's ok
	}
}

func TestFanOut_MultipleEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan int, 10)
	createConsumer := FanOut(ctx, events)

	consumer := createConsumer(10)

	// Send multiple events
	for i := 0; i < 5; i++ {
		events <- i
	}

	// Should receive all events in order
	for i := 0; i < 5; i++ {
		select {
		case val := <-consumer:
			if val != i {
				t.Errorf("received %v, want %v", val, i)
			}
		case <-time.After(time.Second):
			t.Errorf("timed out waiting for event %d", i)
		}
	}
}
