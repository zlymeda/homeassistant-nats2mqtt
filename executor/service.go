package executor

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

func New(ctx context.Context, origin Origin, nc *nats.Conn) *Service {
	return &Service{
		ctx:     ctx,
		nc:      nc,
		origin:  origin,
		devices: map[string]*EntityRegistry{},
		updates: make(chan struct{}, 1),
	}
}

type Service struct {
	ctx context.Context
	nc  *nats.Conn

	origin Origin

	mutex sync.Mutex

	devices map[string]*EntityRegistry

	updates chan struct{}
}

func (s *Service) Start() {
	retry := make(chan struct{}, 1)

	doRetry := func() {
		select {
		case retry <- struct{}{}:
		default:
		}
	}

	publishStates := func() {
		if err := s.PublishStates(); err != nil {
			slog.Error("publishing states", slog.Any("err", err))
			doRetry()
		}
	}

	publishDiscovery := func() {
		for i := range 10 {
			if err := s.PublishDiscovery(); err != nil {
				slog.Error("publishing discovery", slog.Any("err", err), slog.Int("attempt", i+1))
				microSleep()
				continue
			}

			break
		}

		// give HA time to subscribe on the topics
		time.Sleep(5 * time.Second)

		publishStates()
	}

	_, _ = s.nc.Subscribe("homeassistant.status", func(msg *nats.Msg) {
		status := string(msg.Data)
		if status != "online" {
			return
		}

		// Run in a separate goroutine to avoid blocking the NATS message handler.
		go publishDiscovery()
	})

	publishDiscovery()

	for {
		select {
		case <-s.ctx.Done():
			return

		case <-time.After(5 * time.Minute):
			publishStates()

		case <-retry:
			microSleep()
			publishStates()

		case <-s.updates:
			publishDiscovery()
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
		errs = errors.Join(errs, dev.PublishStates())
	}

	return errs
}

func microSleep() {
	time.Sleep((400 + time.Duration(rand.IntN(300))) * time.Millisecond)
}
