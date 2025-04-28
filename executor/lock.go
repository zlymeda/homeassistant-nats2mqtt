package executor

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

func (s *EntityRegistry) AddLock(e *entity.Lock) error {
	lockMeta := observable.NewMapped(e.Meta, func(i entity.LockMeta) entity.Metadata {
		return i
	})

	s.register(entity.Entity{
		Meta: lockMeta,
		State: observable.NewMapped(e.State, func(i entity.LockState) string {
			return string(i)
		}),
		Attributes: e.Attributes,
	})

	commands := map[string]func() error{
		"LOCK":   e.Lock,
		"UNLOCK": e.Unlock,
	}

	meta := lockMeta.Current()

	cmdTopic := s.fullTopic(meta, entity.CommandTopic)
	if err := s.monitorCommandsOn(cmdTopic, createCallback(commands)); err != nil {
		return err
	}

	return nil
}
