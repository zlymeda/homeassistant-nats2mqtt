package executor

import (
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

func (s *EntityRegistry) monitorCommandsOn(cmdTopic string, callback func([]byte) error) error {
	sub, err := s.nc.Subscribe(cmdTopic, func(msg *nats.Msg) {
		cmd := string(msg.Data)

		slog.Info("command", slog.String("topic", cmdTopic), slog.String("cmd", cmd))

		go func(msg *nats.Msg) {
			err := callback(msg.Data)
			if err != nil {
				slog.Warn("command failed",
					slog.String("cmd", cmd),
					slog.String("type", cmdTopic),
					slog.Any("error", err),
				)
			}
		}(msg)

	})

	if err != nil {
		return err
	}

	go func() {
		<-s.ctx.Done()
		_ = sub.Unsubscribe()
	}()

	return nil
}

func createCallback(entityType string, callback map[string]func() error) func([]byte) error {
	return func(body []byte) error {
		cmd := string(body)
		fce, ok := callback[cmd]
		if !ok {
			slog.Warn("unknown mqtt command",
				slog.String("cmd", cmd),
				slog.String("type", entityType),
			)
			return fmt.Errorf("unknown mqtt command: %s", cmd)
		}
		if fce == nil {
			slog.Warn("mqtt command not bound",
				slog.String("cmd", cmd),
				slog.String("type", entityType),
			)
			return fmt.Errorf("mqtt command not bound: %s", cmd)
		}
		return fce()
	}
}
