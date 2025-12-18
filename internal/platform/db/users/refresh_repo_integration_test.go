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

func TestRefreshRepoLifecycle(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock error: %v", err)
	}
	defer db.Close()

	repo := NewRefreshRepo(db)
	now := time.Unix(0, 0).UTC()
	token := domain.NewRefreshTokenRecord("user", "hash", now, time.Hour)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO auth_refresh_tokens")).
		WithArgs(token.ID, token.UserID.String(), token.TokenHash, token.ExpiresAt, token.RevokedAt, token.CreatedAt, nil, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Create(context.Background(), token); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked_at", "created_at", "user_agent", "ip"}).
		AddRow(token.ID, token.UserID.String(), token.TokenHash, token.ExpiresAt, nil, token.CreatedAt, "", "")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT\n            id::text,")).
		WithArgs(token.TokenHash).
		WillReturnRows(rows)

	got, found, err := repo.GetByHash(context.Background(), token.TokenHash)
	if err != nil || !found {
		t.Fatalf("expected token found, err=%v found=%v", err, found)
	}
	if got.TokenHash != token.TokenHash {
		t.Fatalf("unexpected token returned")
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE auth_refresh_tokens")).
		WithArgs(token.ID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Revoke(context.Background(), token.ID); err != nil {
		t.Fatalf("revoke failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
