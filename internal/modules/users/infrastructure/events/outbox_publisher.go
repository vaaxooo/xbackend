package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	userevents "github.com/vaaxooo/xbackend/internal/modules/users/application/events"
)

const userRegisteredEventType = "users.user_registered"
const emailConfirmationRequestedType = "users.email_confirmation_requested"
const passwordResetRequestedType = "users.password_reset_requested"

// OutboxPublisher converts typed application events into raw outbox messages so
// they can be dispatched asynchronously by a background worker.
type OutboxPublisher struct {
	repo *OutboxRepository
}

func NewOutboxPublisher(repo *OutboxRepository) *OutboxPublisher {
	return &OutboxPublisher{repo: repo}
}

func (p *OutboxPublisher) PublishUserRegistered(ctx context.Context, event userevents.UserRegistered) error {
        payload, err := json.Marshal(event)
        if err != nil {
                return err
        }

	return p.repo.Add(ctx, OutboxMessage{
		ID:         uuid.New(),
		EventType:  eventType,
		Payload:    payload,
                OccurredAt: event.OccurredAt,
        })
}

func (p *OutboxPublisher) PublishEmailConfirmationRequested(ctx context.Context, event userevents.EmailConfirmationRequested) error {
        payload, err := json.Marshal(event)
        if err != nil {
                return err
        }

        return p.repo.Add(ctx, OutboxMessage{
                ID:         uuid.New(),
                EventType:  emailConfirmationRequestedType,
                Payload:    payload,
                OccurredAt: event.OccurredAt,
        })
}

func (p *OutboxPublisher) PublishPasswordResetRequested(ctx context.Context, event userevents.PasswordResetRequested) error {
        payload, err := json.Marshal(event)
        if err != nil {
                return err
        }

        return p.repo.Add(ctx, OutboxMessage{
                ID:         uuid.New(),
                EventType:  passwordResetRequestedType,
                Payload:    payload,
                OccurredAt: event.OccurredAt,
        })
}
