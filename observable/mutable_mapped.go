package observable

import (
	"log/slog"
)

var _ Mutable[string] = &mutableMapped[int64, string]{}

// NewMutableMapped creates a Mutable[O] that wraps a Mutable[I] delegate,
// mapping values through the provided mapper/inverseMapper functions.
func NewMutableMapped[I any, O any](delegate Mutable[I], mapper func(I) O, inverseMapper func(O) I) Mutable[O] {
	return &mutableMapped[I, O]{
		delegate:      delegate,
		mapper:        mapper,
		inverseMapper: inverseMapper,
	}
}

type mutableMapped[I any, O any] struct {
	delegate      Mutable[I]
	mapper        func(I) O
	inverseMapper func(O) I
}

func (m *mutableMapped[I, O]) Current() O {
	return m.mapper(m.delegate.Current())
}

func (m *mutableMapped[I, O]) Mutate(state O) {
	m.delegate.Mutate(m.inverseMapper(state))
}

func (m *mutableMapped[I, O]) Changes() <-chan O {
	in := m.delegate.Changes()
	if in == nil {
		return nil
	}

	out := make(chan O, 1)

	go func() {
		defer close(out)
		for v := range in {
			select {
			case out <- m.mapper(v):
			default:
				slog.Debug("mutableMapped: discarding event as the channel is full")
			}
		}
	}()

	return out
}
