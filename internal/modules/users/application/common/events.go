package common

import (
	"context"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/events"
)

// EventPublisher defines outbound integration events produced by the Users bounded context.
// It allows swapping implementations (logger, broker, outbox) without touching use-cases.
type EventPublisher interface {
	PublishUserRegistered(ctx context.Context, event events.UserRegistered) error
	PublishEmailConfirmationRequested(ctx context.Context, event events.EmailConfirmationRequested) error
	PublishPasswordResetRequested(ctx context.Context, event events.PasswordResetRequested) error
}

// NopEventPublisher is useful for tests or environments where outbound delivery is disabled.
type NopEventPublisher struct{}

func (NopEventPublisher) PublishUserRegistered(_ context.Context, _ events.UserRegistered) error {
	return nil
}

func (NopEventPublisher) PublishEmailConfirmationRequested(_ context.Context, _ events.EmailConfirmationRequested) error {
	return nil
}

func (NopEventPublisher) PublishPasswordResetRequested(_ context.Context, _ events.PasswordResetRequested) error {
	return nil
}
