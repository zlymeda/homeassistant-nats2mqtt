package fanner

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

func FanOut[E any](ctx context.Context, events <-chan E) func(buffer int) <-chan E {
	var consumers []chan E
	var mutex sync.Mutex

	go func() {
		defer func() {
			mutex.Lock()
			for _, ch := range consumers {
				close(ch)
			}
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

				mutex.Lock()
				activeConsumers := make([]chan E, 0, len(consumers))
				for _, ch := range consumers {
					select {
					case ch <- event:
						activeConsumers = append(activeConsumers, ch)
					case <-time.After(1 * time.Second):
						// If the consumer channel is full, discard the event
						slog.Warn("fanout: discarding event as the channel is full",
							slog.Any("event", event))
					}
				}

				consumers = activeConsumers
				mutex.Unlock()
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
