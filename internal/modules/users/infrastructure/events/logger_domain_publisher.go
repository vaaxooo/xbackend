package events

import (
	"context"
	"sync"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	plog "github.com/vaaxooo/xbackend/internal/platform/log"
)

// LoggerDomainPublisher sends raw outbox messages to the configured logger.
// It keeps track of already published IDs to avoid duplicate logs when the
// worker retries the same event.
type LoggerDomainPublisher struct {
	logger    plog.Logger
	processed sync.Map
}

func NewLoggerDomainPublisher(logger plog.Logger) *LoggerDomainPublisher {
	return &LoggerDomainPublisher{logger: logger}
}

func (p *LoggerDomainPublisher) Publish(ctx context.Context, eventID string, eventType string, payload []byte) error {
	if _, loaded := p.processed.LoadOrStore(eventID, struct{}{}); loaded {
		p.logger.Debug(ctx, "outbox event already published", "event_id", eventID, "event_type", eventType)
		return nil
	}

	p.logger.Info(ctx, "outbox event published", "event_id", eventID, "event_type", eventType, "payload", string(payload))
	return nil
}

var _ common.DomainEventPublisher = (*LoggerDomainPublisher)(nil)
