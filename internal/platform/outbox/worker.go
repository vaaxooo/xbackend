package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	userevents "github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/events"
	plog "github.com/vaaxooo/xbackend/internal/platform/log"
)

// Repository captures the persistence operations needed by the worker.
type Repository interface {
	GetPending(ctx context.Context, limit int) ([]userevents.OutboxMessage, error)
	MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time) error
	MarkFailed(ctx context.Context, id uuid.UUID, publishErr error) error
}

// Config configures worker behaviour.
type Config struct {
	BatchSize int
	Interval  time.Duration
}

// Worker replays outbox records to the configured transport.
type Worker struct {
	repo      Repository
	publisher common.DomainEventPublisher
	logger    plog.Logger
	clock     func() time.Time
	cfg       Config
}

func NewWorker(repo Repository, publisher common.DomainEventPublisher, logger plog.Logger, cfg Config) *Worker {
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 32
	}
	if cfg.Interval == 0 {
		cfg.Interval = 5 * time.Second
	}
	return &Worker{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
		clock:     time.Now,
		cfg:       cfg,
	}
}

// Run launches a loop that periodically flushes pending outbox messages.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()

	for {
		if err := w.ProcessOnce(ctx); err != nil {
			w.logger.Error(ctx, "outbox worker iteration failed", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// ProcessOnce fetches a batch of pending events and attempts to deliver them.
func (w *Worker) ProcessOnce(ctx context.Context) error {
	messages, err := w.repo.GetPending(ctx, w.cfg.BatchSize)
	if err != nil {
		return err
	}

	for _, msg := range messages {
		if err := w.publisher.Publish(ctx, msg.ID.String(), msg.EventType, msg.Payload); err != nil {
			if markErr := w.repo.MarkFailed(ctx, msg.ID, err); markErr != nil {
				w.logger.Error(ctx, "failed to mark outbox message as failed", markErr, "event_id", msg.ID.String())
			}
			w.logger.Warn(ctx, "outbox publish failed", "event_id", msg.ID.String(), "event_type", msg.EventType)
			continue
		}

		if err := w.repo.MarkPublished(ctx, msg.ID, w.clock().UTC()); err != nil {
			w.logger.Error(ctx, "failed to mark outbox message as published", err, "event_id", msg.ID.String())
		}
	}

	return nil
}
