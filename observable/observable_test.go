package observable

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewSimple(t *testing.T) {
	ctx := context.Background()
	s := NewSimple[string](ctx, "initial")

	if got := s.Current(); got != "initial" {
		t.Errorf("Current() = %v, want %v", got, "initial")
	}
}

func TestSimple_Change(t *testing.T) {
	ctx := context.Background()
	s := NewSimple[string](ctx, "initial")

	s.Change("updated")

	if got := s.Current(); got != "updated" {
		t.Errorf("Current() = %v, want %v", got, "updated")
	}
}

func TestSimple_Mutate(t *testing.T) {
	ctx := context.Background()
	s := NewSimple[string](ctx, "initial")

	s.Mutate("mutated")

	if got := s.Current(); got != "mutated" {
		t.Errorf("Current() = %v, want %v", got, "mutated")
	}
}

func TestSimple_Changes(t *testing.T) {
	ctx := context.Background()
	s := NewSimple[string](ctx, "initial")

	changes := s.Changes()
	if changes == nil {
		t.Fatal("Changes() returned nil")
	}

	// Change the value
	s.Change("updated")

	// Should receive the change
	select {
	case val := <-changes:
		if val != "updated" {
			t.Errorf("received %v, want %v", val, "updated")
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for change")
	}
}

func TestNewSingle(t *testing.T) {
	s := NewSingle[string]("fixed")

	if got := s.Current(); got != "fixed" {
		t.Errorf("Current() = %v, want %v", got, "fixed")
	}
}

func TestSingle_Changes(t *testing.T) {
	s := NewSingle[string]("fixed")

	changes := s.Changes()
	if changes != nil {
		t.Error("Changes() should return nil for Single")
	}
}

func TestNewMapped(t *testing.T) {
	ctx := context.Background()
	inner := NewSimple[int](ctx, 42)
	mapped := NewMapped(inner, func(i int) string {
		return fmt.Sprintf("%d", i)
	})

	if got := mapped.Current(); got != "42" {
		t.Errorf("Current() = %v, want %v", got, "42")
	}
}

func TestMapped_Changes(t *testing.T) {
	ctx := context.Background()
	inner := NewSimple[int](ctx, 42)
	mapped := NewMapped(inner, func(i int) string {
		return fmt.Sprintf("%d", i)
	})

	changes := mapped.Changes()
	if changes == nil {
		t.Fatal("Changes() returned nil")
	}

	// Change the inner value
	inner.Change(100)

	// Should receive the mapped change
	select {
	case val := <-changes:
		if val != "100" {
			t.Errorf("received %v, want %v", val, "100")
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for change")
	}
}

func TestMapped_NilChanges(t *testing.T) {
	inner := NewSingle[int](42)
	mapped := NewMapped(inner, func(i int) string {
		return fmt.Sprintf("%d", i)
	})

	changes := mapped.Changes()
	if changes != nil {
		t.Error("Changes() should return nil when inner has nil changes")
	}
}
