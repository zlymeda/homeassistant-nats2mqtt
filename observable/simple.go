package observable

import (
	"context"
	"log/slog"
	"sync"

	"github.com/zlymeda/homeassistant-nats2mqtt/internal/fanner"
)

var _ Observable[string] = &Simple[string]{}

func NewSimple[T any](ctx context.Context, state T) *Simple[T] {
	states := make(chan T, 1)
	return &Simple[T]{
		state:          state,
		states:         states,
		createListener: fanner.FanOut(ctx, states),
	}
}

type Simple[T any] struct {
	mutex sync.RWMutex

	state          T
	states         chan T
	createListener func(buffer int) <-chan T
}

func (s *Simple[T]) Change(state T) {
	s.mutex.Lock()
	s.state = state
	s.mutex.Unlock()

	// Send on channel without holding the mutex to avoid deadlock.
	// Non-blocking: if the fan-out goroutine hasn't drained the previous
	// event yet, we discard this one (consumers will get the latest via Current()).
	select {
	case s.states <- state:
	default:
		slog.Debug("observable: discarding intermediate state change, fan-out buffer full")
	}
}

func (s *Simple[T]) Current() T {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.state
}

func (s *Simple[T]) Changes() <-chan T {
	return s.createListener(1)
}
