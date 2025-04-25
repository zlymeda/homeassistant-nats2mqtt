package observable

var _ Observable[string] = &Single[string]{}

func NewSingle[T any](state T) *Single[T] {
	return &Single[T]{
		state: state,
	}
}

type Single[T any] struct {
	state T
}

func (s Single[T]) Current() T {
	return s.state
}

func (s Single[T]) Changes() <-chan T {
	return nil
}
