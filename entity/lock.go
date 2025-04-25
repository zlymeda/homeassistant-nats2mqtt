package entity

import (
	"github.com/zlymeda/homeassistant-nats2mqtt/observable"
)

var _ Metadata = LockMeta{}

type LockMeta struct {
	Meta
}

func (l LockMeta) GetId() string {
	return l.Id
}

func (l LockMeta) GetName() string {
	return l.Name
}

func (l LockMeta) GetExtraTopics() []string {
	return []string{CommandTopic}
}

func (l LockMeta) ToHaDiscovery(dev Device) map[string]any {
	return l.Meta.ToHaDiscovery(dev, "lock")
}

type LockState string

const (
	LockStateLocked   LockState = "LOCKED"
	LockStateUnlocked LockState = "UNLOCKED"
)

type Lock struct {
	Meta       observable.Observable[Meta]
	State      observable.Observable[LockState]
	Attributes observable.Observable[Attrs]

	Lock   func() error
	Unlock func() error
}
