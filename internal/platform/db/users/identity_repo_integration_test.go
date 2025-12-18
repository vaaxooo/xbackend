//go:build integration

package usersdb

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

func TestIdentityRepoCreateAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock error: %v", err)
	}
	defer db.Close()

	repo := NewIdentityRepo(db)
	identity := domain.NewEmailIdentity("user", mustEmail(t, "user@example.com"), "hash", time.Unix(0, 0))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO auth_identities")).
		WithArgs(identity.ID, identity.UserID.String(), identity.Provider, identity.ProviderUserID, identity.SecretHash.String(), identity.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Create(context.Background(), identity); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_user_id", "secret_hash", "created_at"}).
		AddRow(identity.ID, identity.UserID.String(), identity.Provider, identity.ProviderUserID, "hash", identity.CreatedAt)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT\n            id::text,")).
		WithArgs(identity.Provider, identity.ProviderUserID).
		WillReturnRows(rows)

	got, found, err := repo.GetByProvider(context.Background(), identity.Provider, identity.ProviderUserID)
	if err != nil || !found {
		t.Fatalf("expected identity found, err=%v found=%v", err, found)
	}
	if got.Provider != identity.Provider || got.UserID != identity.UserID {
		t.Fatalf("unexpected identity: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func mustEmail(t *testing.T, raw string) domain.Email {
	t.Helper()
	e, err := domain.NewEmail(raw)
	if err != nil {
		t.Fatalf("email error: %v", err)
	}
	return e
}
