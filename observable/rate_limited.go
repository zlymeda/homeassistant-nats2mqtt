package observable

import (
	"context"
	"log/slog"
	"reflect"
	"sync"
	"time"
)

var _ Observable[string] = &RateLimited[string]{}

type RateLimitOptions[T any] struct {
	MinInterval time.Duration
	Equal       func(a, b T) bool
}

func NewRateLimited[T any](ctx context.Context, delegate Observable[T], opts RateLimitOptions[T]) *RateLimited[T] {
	if opts.Equal == nil {
		opts.Equal = func(a, b T) bool {
			return reflect.DeepEqual(a, b)
		}
	}

	r := &RateLimited[T]{
		ctx:         ctx,
		delegate:    delegate,
		minInterval: opts.MinInterval,
		equal:       opts.Equal,
		state:       delegate.Current(),
		listeners:   map[chan T]struct{}{},
	}
	go r.monitor()
	return r
}

type RateLimited[T any] struct {
	ctx      context.Context
	delegate Observable[T]

	minInterval time.Duration
	equal       func(a, b T) bool

	mu         sync.RWMutex
	state      T
	lastEmit   time.Time
	timer      *time.Timer
	pending    T
	hasPending bool
	listeners  map[chan T]struct{}
}

func (r *RateLimited[T]) Current() T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

func (r *RateLimited[T]) Changes() <-chan T {
	ch := make(chan T, 1)

	r.mu.Lock()
	r.listeners[ch] = struct{}{}
	r.mu.Unlock()

	go func() {
		<-r.ctx.Done()
		r.mu.Lock()
		delete(r.listeners, ch)
		close(ch)
		r.mu.Unlock()
	}()

	return ch
}

func (r *RateLimited[T]) monitor() {
	in := r.delegate.Changes()
	if in == nil {
		return
	}

	for {
		select {
		case <-r.ctx.Done():
			if r.timer != nil {
				r.timer.Stop()
			}
			return

		case value, ok := <-in:
			if !ok {
				return
			}
			r.accept(value)
		}
	}
}

func (r *RateLimited[T]) accept(value T) {
	r.mu.Lock()
	if r.equal(r.state, value) {
		r.mu.Unlock()
		return
	}

	now := time.Now()
	if r.minInterval <= 0 || r.lastEmit.IsZero() || now.Sub(r.lastEmit) >= r.minInterval {
		r.emitLocked(value)
		r.mu.Unlock()
		return
	}

	r.pending = value
	r.hasPending = true
	if r.timer == nil {
		r.timer = time.AfterFunc(r.minInterval-now.Sub(r.lastEmit), r.flush)
	}
	r.mu.Unlock()
}

func (r *RateLimited[T]) flush() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.timer = nil
	if !r.hasPending {
		return
	}

	value := r.pending
	var zero T
	r.pending = zero
	r.hasPending = false

	if r.equal(r.state, value) {
		return
	}
	r.emitLocked(value)
}

func (r *RateLimited[T]) emitLocked(value T) {
	r.state = value
	r.lastEmit = time.Now()

	for listener := range r.listeners {
		select {
		case listener <- value:
		default:
			slog.Debug("rateLimitedObservable: discarding event as the channel is full")
		}
	}
}
