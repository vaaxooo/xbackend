package events

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	pdb "github.com/vaaxooo/xbackend/internal/platform/db"
)

// OutboxRepository persists integration events in the same transaction as the
// application use-case. It also exposes helpers for background delivery.
type OutboxRepository struct {
	db    *sql.DB
	clock func() time.Time
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db, clock: time.Now}
}

// Add stores a new outbox record. If the ID is zero, a new UUID is generated.
func (r *OutboxRepository) Add(ctx context.Context, msg OutboxMessage) error {
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	now := r.clock().UTC()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}

	const query = `
                INSERT INTO user_events_outbox (id, event_type, payload, occurred_at, created_at)
                VALUES ($1::uuid, $2, $3, $4, $5)
        `

	exec := pdb.Executor(ctx, r.db)
	_, err := exec.ExecContext(ctx, query, msg.ID.String(), msg.EventType, msg.Payload, msg.OccurredAt, msg.CreatedAt)
	return err
}

// GetPending selects events that were not published yet ordered by creation time.
func (r *OutboxRepository) GetPending(ctx context.Context, limit int) ([]OutboxMessage, error) {
	if limit <= 0 {
		limit = 32
	}

	const query = `
                SELECT id::text, event_type, payload, occurred_at, created_at, published_at, attempts, last_error
                FROM user_events_outbox
                WHERE published_at IS NULL
                ORDER BY created_at
                LIMIT $1
        `

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []OutboxMessage
	for rows.Next() {
		var m OutboxMessage
		var id string
		var publishedAt sql.NullTime
		var lastError sql.NullString

		if err := rows.Scan(&id, &m.EventType, &m.Payload, &m.OccurredAt, &m.CreatedAt, &publishedAt, &m.Attempts, &lastError); err != nil {
			return nil, err
		}
		parsedID, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid outbox id %q: %w", id, err)
		}
		m.ID = parsedID
		if publishedAt.Valid {
			m.PublishedAt = &publishedAt.Time
		}
		if lastError.Valid {
			msg := lastError.String
			m.LastError = &msg
		}

		out = append(out, m)
	}

	return out, rows.Err()
}

// MarkPublished records a successful publication attempt. Attempts counter is
// incremented alongside the published timestamp to make retries idempotent.
func (r *OutboxRepository) MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time) error {
	const query = `
                UPDATE user_events_outbox
                SET published_at = $2, attempts = attempts + 1, last_error = NULL
                WHERE id = $1::uuid
        `
	_, err := r.db.ExecContext(ctx, query, id.String(), publishedAt.UTC())
	return err
}

// MarkFailed increments attempts counter and stores last error for observability.
func (r *OutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, publishErr error) error {
	const query = `
                UPDATE user_events_outbox
                SET attempts = attempts + 1, last_error = $2
                WHERE id = $1::uuid
        `

	var errMsg any
	if publishErr != nil {
		errMsg = publishErr.Error()
	}

	res, err := r.db.ExecContext(ctx, query, id.String(), errMsg)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("outbox record not found")
	}
	return nil
}

// Ensure the repository can be used by the outbox worker.
var _ interface {
	GetPending(context.Context, int) ([]OutboxMessage, error)
	MarkPublished(context.Context, uuid.UUID, time.Time) error
	MarkFailed(context.Context, uuid.UUID, error) error
} = (*OutboxRepository)(nil)
