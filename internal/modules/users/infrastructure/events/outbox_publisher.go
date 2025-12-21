package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	userevents "github.com/vaaxooo/xbackend/internal/modules/users/application/events"
)

// OutboxPublisher converts typed application events into raw outbox messages so
// they can be dispatched asynchronously by a background worker.
type outboxWriter interface {
	Add(ctx context.Context, msg OutboxMessage) error
}

type OutboxPublisher struct {
	repo outboxWriter
}

func NewOutboxPublisher(repo outboxWriter) *OutboxPublisher {
	return &OutboxPublisher{repo: repo}
}

func (p *OutboxPublisher) PublishUserRegistered(ctx context.Context, event userevents.UserRegistered) error {
	return p.publish(ctx, EventTypeUserRegistered, event.OccurredAt, event)
}

func (p *OutboxPublisher) PublishEmailConfirmationRequested(ctx context.Context, event userevents.EmailConfirmationRequested) error {
	return p.publish(ctx, EventTypeEmailConfirmationRequested, event.OccurredAt, event)
}

func (p *OutboxPublisher) PublishPasswordResetRequested(ctx context.Context, event userevents.PasswordResetRequested) error {
	return p.publish(ctx, EventTypePasswordResetRequested, event.OccurredAt, event)
}

// Ensure OutboxPublisher conforms to application contract.
var _ common.EventPublisher = (*OutboxPublisher)(nil)

func (p *OutboxPublisher) publish(ctx context.Context, eventType string, occurredAt time.Time, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.repo.Add(ctx, OutboxMessage{
		ID:         uuid.New(),
		EventType:  eventType,
		Payload:    payload,
		OccurredAt: occurredAt,
	})
}
