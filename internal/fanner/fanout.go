package fanner

import (
	"context"
	"log/slog"
	"sync"
)

func FanOut[E any](ctx context.Context, events <-chan E) func(buffer int) <-chan E {
	var consumers []chan E
	var mutex sync.RWMutex

	go func() {
		defer func() {
			mutex.Lock()
			for _, ch := range consumers {
				close(ch)
			}
			consumers = nil
			mutex.Unlock()
		}()

		for {
			select {
			case <-ctx.Done():
				slog.Info("fanout: ctx is done; cancelling consumers distribution")
				return

			case event, ok := <-events:
				if !ok {
					// Input channel was closed
					return
				}

				// Take a snapshot of consumers under a read lock to minimize lock hold time.
				mutex.RLock()
				snapshot := make([]chan E, len(consumers))
				copy(snapshot, consumers)
				mutex.RUnlock()

				// Distribute to consumers without holding the lock.
				// Use non-blocking send: discard event for slow consumers
				// but keep the consumer registered (don't evict).
				for _, ch := range snapshot {
					select {
					case ch <- event:
					default:
						slog.Warn("fanout: discarding event for slow consumer",
							slog.Any("event", event))
					}
				}
			}
		}
	}()

	return func(buffer int) <-chan E {
		ch := make(chan E, buffer)

		mutex.Lock()
		defer mutex.Unlock()

		consumers = append(consumers, ch)

		return ch
	}
}
