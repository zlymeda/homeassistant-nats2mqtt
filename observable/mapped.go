package observable

import (
	"log/slog"
)

var _ Observable[string] = &mappedObservable[int64, string]{}

func NewMapped[I any, O any](delegate Observable[I], mapper func(I) O) Observable[O] {
	return &mappedObservable[I, O]{
		delegate: delegate,
		mapper:   mapper,
	}
}

type mappedObservable[I any, O any] struct {
	delegate Observable[I]
	mapper   func(I) O
}

func (m *mappedObservable[I, O]) Current() O {
	return m.mapper(m.delegate.Current())
}

func (m *mappedObservable[I, O]) Changes() <-chan O {
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
				slog.Debug("mappedObservable: discarding event as the channel is full")
			}
		}
	}()

	return out
}
