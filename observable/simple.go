package observable

import (
	"context"
	"github.com/zlymeda/homeassistant-nats2mqtt/internal/fanner"
	"sync"
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
	mutex sync.Mutex

	state          T
	states         chan T
	createListener func(buffer int) <-chan T
}

func (s *Simple[T]) Change(state T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.state = state
	s.states <- state
}

func (s *Simple[T]) Current() T {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.state
}

func (s *Simple[T]) Changes() <-chan T {
	return s.createListener(1)
}
