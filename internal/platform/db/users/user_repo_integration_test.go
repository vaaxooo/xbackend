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

func TestUserRepoCreateAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewUserRepo(db)
	displayName, err := domain.NewDisplayName("Display")
	if err != nil {
		t.Fatalf("failed to create display name: %v", err)
	}
	user := domain.NewUser(domain.UserID("user-1"), displayName, time.Unix(0, 0))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
		WithArgs(user.ID.String(), sqlmock.AnyArg(), nil, user.ProfileCustomized, user.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "middle_name", "display_name", "avatar_url", "profile_customized", "created_at"}).
		AddRow("user-1", "", "", "", "Display", "", false, user.CreatedAt)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT\n                        id::text")).
		WithArgs(user.ID.String()).
		WillReturnRows(rows)

	got, found, err := repo.GetByID(context.Background(), user.ID)
	if err != nil || !found {
		t.Fatalf("expected user found, got err=%v found=%v", err, found)
	}
	if got.DisplayName != "Display" || got.ID != user.ID {
		t.Fatalf("unexpected user data: %+v", got)
	}

	updateRows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "middle_name", "display_name", "avatar_url", "profile_customized", "created_at"}).
		AddRow("user-1", "John", "", "", "Display", "", true, user.CreatedAt)
	mock.ExpectQuery(regexp.QuoteMeta("UPDATE users")).WillReturnRows(updateRows)

	patchedUser, err := got.ApplyPatch(domain.ProfilePatch{FirstName: ptr("John")})
	if err != nil {
		t.Fatalf("apply patch failed: %v", err)
	}
	patched, err := repo.UpdateProfile(context.Background(), patchedUser)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if patched.FirstName != "John" || !patched.ProfileCustomized {
		t.Fatalf("unexpected updated user: %+v", patched)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func ptr[T any](v T) *T { return &v }
