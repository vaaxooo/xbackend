//go:build integration

package register

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/events"
	pdb "github.com/vaaxooo/xbackend/internal/platform/db"
	usersdb "github.com/vaaxooo/xbackend/internal/platform/db/users"
)

func TestRegisterUseCaseWritesOutboxInsideTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock error: %v", err)
	}
	defer db.Close()

	uow := pdb.NewUnitOfWork(db)
	usersRepo := usersdb.NewUserRepo(db)
	identitiesRepo := usersdb.NewIdentityRepo(db)
	refreshRepo := usersdb.NewRefreshRepo(db)
	outboxRepo := events.NewOutboxRepository(db)

	publisher := events.NewOutboxPublisher(outboxRepo)
	uc := common.NewTransactionalUseCase(uow, New(usersRepo, identitiesRepo, refreshRepo, stubHasher{}, stubTokenIssuer{}, publisher, time.Minute, time.Hour))

	mock.ExpectQuery(`SELECT\s+id::text,\s+user_id::text,\s+provider,\s+provider_user_id,\s+COALESCE\(secret_hash, ''\),\s+created_at\s+FROM auth_identities\s+WHERE provider = \$1 AND provider_user_id = \$2\s+LIMIT 1`).
		WithArgs("email", "john@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_user_id", "secret_hash", "created_at"}))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO auth_identities")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "email", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO auth_refresh_tokens")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO user_events_outbox")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if _, err := uc.Execute(context.Background(), Input{Email: "john@example.com", Password: "strongpassword", DisplayName: "John"}); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
