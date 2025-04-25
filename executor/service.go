package executor

import (
	"context"
	"errors"
	"github.com/nats-io/nats.go"
	"log/slog"
	"math/rand"

	"sync"
	"time"
)

func New(ctx context.Context, origin Origin, nc *nats.Conn) *Service {
	return &Service{
		ctx:     ctx,
		nc:      nc,
		origin:  origin,
		devices: map[string]*EntityRegistry{},
	}
}

type Service struct {
	ctx context.Context
	nc  *nats.Conn

	origin Origin

	mutex sync.Mutex

	devices map[string]*EntityRegistry
}

func (s *Service) Start() {
	retry := make(chan struct{})

	doRetry := func() {
		select {
		case <-s.ctx.Done():
			return

		case retry <- struct{}{}:

		default:
			return
		}

	}

	publish := func() {
		if err := s.PublishStates(); err != nil {
			slog.Error("publishing states", slog.Any("err", err))
			doRetry()
		}
	}

	_, _ = s.nc.Subscribe("homeassistant.status", func(msg *nats.Msg) {
		status := string(msg.Data)
		if status != "online" {
			return
		}

		for i := 0; i < 10; i++ {
			if err := s.PublishDiscovery(); err != nil {
				slog.Error("publishing discovery", slog.Any("err", err))
				microSleep()
				continue
			}

			break
		}

		// give HA time to subscribe on the topics
		time.Sleep(5 * time.Second)

		publish()
	})

	publish()

	for {
		select {
		case <-s.ctx.Done():
			return

		case <-time.After(5 * time.Minute):
			publish()

		case <-retry:
			microSleep()
			publish()
		}
	}
}

func (s *Service) PublishDiscovery() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var errs error
	for _, dev := range s.devices {
		err := dev.PublishDiscovery()
		errs = errors.Join(errs, err)
	}

	return errs
}

func (s *Service) PublishStates() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var errs error
	for _, dev := range s.devices {
		for _, publish := range dev.publishState {
			errs = errors.Join(errs, publish())
		}
	}

	return errs
}

func microSleep() {
	time.Sleep((400 + time.Duration(rand.Int31n(300))) * time.Millisecond)
}
