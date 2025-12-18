package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	userevents "github.com/vaaxooo/xbackend/internal/modules/users/application/events"
	plog "github.com/vaaxooo/xbackend/internal/platform/log"
)

// LoggerPublisher is a minimal integration-event publisher that writes payloads to the configured logger.
// It keeps the application layer decoupled from the transport and can later be swapped with Kafka/NATS/etc.
type LoggerPublisher struct {
	logger plog.Logger
	clock  func() time.Time
}

func NewLoggerPublisher(logger plog.Logger) *LoggerPublisher {
	return &LoggerPublisher{
		logger: logger,
		clock:  time.Now,
	}
}

func (p *LoggerPublisher) PublishUserRegistered(ctx context.Context, event userevents.UserRegistered) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	p.logger.Info(ctx, "user.registered", "event", string(payload), "published_at", p.clock().UTC())
	return nil
}

// Ensure LoggerPublisher implements the application contract.
var _ common.EventPublisher = (*LoggerPublisher)(nil)
