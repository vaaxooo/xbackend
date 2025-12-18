//go:build integration

package outbox

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/events"
	plog "github.com/vaaxooo/xbackend/internal/platform/log"
)

func TestWorkerIsIdempotentAcrossRetries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock error: %v", err)
	}
	defer db.Close()

	repo := events.NewOutboxRepository(db)
	publisher := &recordingPublisher{}
	worker := NewWorker(repo, publisher, plog.New("dev"), Config{})

	eventID := uuid.New().String()
	rows := sqlmock.NewRows([]string{"id", "event_type", "payload", "occurred_at", "created_at", "published_at", "attempts", "last_error"}).
		AddRow(eventID, "test.event", []byte(`{"key":"value"}`), time.Unix(0, 0), time.Unix(0, 0), nil, 0, nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id::text, event_type, payload, occurred_at, created_at, published_at, attempts, last_error\n                FROM user_events_outbox\n                WHERE published_at IS NULL\n                ORDER BY created_at\n                LIMIT $1")).
		WithArgs(32).
		WillReturnRows(rows)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE user_events_outbox\n                SET published_at = $2, attempts = attempts + 1, last_error = NULL\n                WHERE id = $1::uuid\n        ")).
		WithArgs(eventID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id::text, event_type, payload, occurred_at, created_at, published_at, attempts, last_error\n                FROM user_events_outbox\n                WHERE published_at IS NULL\n                ORDER BY created_at\n                LIMIT $1")).
		WithArgs(32).
		WillReturnRows(sqlmock.NewRows([]string{"id", "event_type", "payload", "occurred_at", "created_at", "published_at", "attempts", "last_error"}))

	if err := worker.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("first iteration failed: %v", err)
	}
	if err := worker.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("second iteration failed: %v", err)
	}

	if publisher.calls != 1 {
		t.Fatalf("expected single publish call, got %d", publisher.calls)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

type recordingPublisher struct{ calls int }

func (p *recordingPublisher) Publish(context.Context, string, string, []byte) error {
	p.calls++
	return nil
}
