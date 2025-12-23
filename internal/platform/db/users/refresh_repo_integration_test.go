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

	repo := NewRefreshRepo(db, -1)
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

	listRows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked_at", "created_at", "user_agent", "ip"}).
		AddRow(token.ID, token.UserID.String(), token.TokenHash, token.ExpiresAt, nil, token.CreatedAt, "agent", "1.1.1.1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT\n            id::text,\n            user_id::text,\n            token_hash,\n            expires_at,\n            revoked_at,\n            created_at,\n            COALESCE(user_agent, ''),\n            COALESCE(ip, '')\n        FROM auth_refresh_tokens\n        WHERE user_id = $1::uuid AND revoked_at IS NULL AND expires_at > $2\n        ORDER BY created_at DESC\n        LIMIT 15")).
		WithArgs(token.UserID.String(), sqlmock.AnyArg()).
		WillReturnRows(listRows)

	tokens, err := repo.ListByUser(context.Background(), token.UserID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(tokens) != 1 || tokens[0].UserAgent != "agent" || tokens[0].IP != "1.1.1.1" {
		t.Fatalf("unexpected list result: %+v", tokens)
	}

	getByIDRows := sqlmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "revoked_at", "created_at", "user_agent", "ip"}).
		AddRow(token.ID, token.UserID.String(), token.TokenHash, token.ExpiresAt, nil, token.CreatedAt, "", "")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT\n            id::text,\n            user_id::text,\n            token_hash,\n            expires_at,\n            revoked_at,\n            created_at,\n            COALESCE(user_agent, ''),\n            COALESCE(ip, '')\n        FROM auth_refresh_tokens\n        WHERE id = $1::uuid\n        LIMIT 1")).
		WithArgs(token.ID).
		WillReturnRows(getByIDRows)

	_, found, err = repo.GetByID(context.Background(), token.ID)
	if err != nil || !found {
		t.Fatalf("expected token by id, err=%v found=%v", err, found)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE auth_refresh_tokens")).
		WithArgs(token.ID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Revoke(context.Background(), token.ID); err != nil {
		t.Fatalf("revoke failed: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE auth_refresh_tokens
        SET revoked_at = $3
        WHERE user_id = $1::uuid AND revoked_at IS NULL AND id <> ALL($2::uuid[])`)).
		WithArgs(token.UserID.String(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.RevokeAllExcept(context.Background(), token.UserID, []string{token.ID}); err != nil {
		t.Fatalf("revoke others failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
