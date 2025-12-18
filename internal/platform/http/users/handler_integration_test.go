//go:build integration

package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"

	usersapp "github.com/vaaxooo/xbackend/internal/modules/users/application"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/profile"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/refresh"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/register"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
	"github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/events"
	"github.com/vaaxooo/xbackend/internal/modules/users/public"
	pdb "github.com/vaaxooo/xbackend/internal/platform/db"
	usersdb "github.com/vaaxooo/xbackend/internal/platform/db/users"
	phttp "github.com/vaaxooo/xbackend/internal/platform/http"
	"github.com/vaaxooo/xbackend/internal/platform/log"
)

type stubLogger struct{}

func (stubLogger) Debug(context.Context, string, ...any)        {}
func (stubLogger) Info(context.Context, string, ...any)         {}
func (stubLogger) Warn(context.Context, string, ...any)         {}
func (stubLogger) Error(context.Context, string, error, ...any) {}

// Ensure stubLogger satisfies the interface at compile time.
var _ log.Logger = stubLogger{}

type fakeService struct {
	registerOut login.Output
	registerErr error

	loginOut login.Output
	loginErr error

	refreshOut refresh.Output
	refreshErr error

	getOut profile.Output
	getErr error

	updateOut profile.Output
	updateErr error

	linkOut link.Output
	linkErr error
}

func (f *fakeService) Register(context.Context, register.Input) (login.Output, error) {
	return f.registerOut, f.registerErr
}
func (f *fakeService) Login(context.Context, login.Input) (login.Output, error) {
	return f.loginOut, f.loginErr
}
func (f *fakeService) Refresh(context.Context, refresh.Input) (refresh.Output, error) {
	return f.refreshOut, f.refreshErr
}
func (f *fakeService) GetMe(context.Context, profile.GetInput) (profile.Output, error) {
	return f.getOut, f.getErr
}
func (f *fakeService) UpdateProfile(context.Context, profile.UpdateInput) (profile.Output, error) {
	return f.updateOut, f.updateErr
}
func (f *fakeService) LinkProvider(context.Context, link.Input) (link.Output, error) {
	return f.linkOut, f.linkErr
}

type fakeTokenParser struct {
	userID string
	err    error
}

func (f *fakeTokenParser) Parse(string) (string, error)                { return f.userID, f.err }
func (f *fakeTokenParser) Issue(string, time.Duration) (string, error) { return "token", nil }
func (f *fakeTokenParser) Verify(string) (public.AuthContext, error) {
	if f.err != nil {
		return public.AuthContext{}, f.err
	}
	return public.AuthContext{UserID: f.userID}, nil
}

type stubHasher struct{}

func (stubHasher) Hash(context.Context, string) (string, error)  { return "hash", nil }
func (stubHasher) Compare(context.Context, string, string) error { return nil }

func newTestServer(svc usersapp.Service, tp *fakeTokenParser) *httptest.Server {
	router := phttp.NewRouter(phttp.RouterDeps{Logger: stubLogger{}, Timeout: time.Second}, func(r chi.Router) {
		RegisterV1(r, svc, tp)
	})
	return httptest.NewServer(router)
}

func TestRegisterEndpoint(t *testing.T) {
	svc := &fakeService{registerOut: login.Output{UserID: "id", DisplayName: "User", AccessToken: "acc", RefreshToken: "ref"}}
	server := newTestServer(svc, &fakeTokenParser{})
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "password123", "display_name": "User"})
	resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if out["access_token"] != "acc" || out["refresh_token"] != "ref" {
		t.Fatalf("unexpected tokens: %+v", out)
	}
}

func TestLoginUnauthorized(t *testing.T) {
	svc := &fakeService{loginErr: domain.ErrInvalidCredentials}
	server := newTestServer(svc, &fakeTokenParser{})
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "bad"})
	resp, _ := http.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestProfileEndpoints(t *testing.T) {
	svc := &fakeService{
		getOut:    profile.Output{UserID: "user", DisplayName: "User"},
		updateOut: profile.Output{UserID: "user", FirstName: "New"},
	}
	tp := &fakeTokenParser{userID: "user"}
	server := newTestServer(svc, tp)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer token")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	updateBody, _ := json.Marshal(map[string]string{"first_name": "New"})
	req, _ = http.NewRequest(http.MethodPatch, server.URL+"/api/v1/me", bytes.NewReader(updateBody))
	req.Header.Set("Authorization", "Bearer token")
	resp, _ = http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d", resp.StatusCode)
	}
}

func TestRefreshUnauthorized(t *testing.T) {
	svc := &fakeService{refreshErr: domain.ErrRefreshTokenInvalid}
	server := newTestServer(svc, &fakeTokenParser{})
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"refresh_token": "bad"})
	resp, _ := http.Post(server.URL+"/api/v1/auth/refresh", "application/json", bytes.NewReader(body))
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLinkConflict(t *testing.T) {
	svc := &fakeService{linkErr: domain.ErrIdentityAlreadyLinked}
	tp := &fakeTokenParser{userID: "user"}
	server := newTestServer(svc, tp)
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"provider": "github", "provider_user_id": "gh-1"})
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/auth/link", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer token")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestRegisterEndpointTransactionalCommit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	uow := pdb.NewUnitOfWork(db)
	usersRepo := usersdb.NewUserRepo(db)
	identitiesRepo := usersdb.NewIdentityRepo(db)
	refreshRepo := usersdb.NewRefreshRepo(db)
	outboxRepo := events.NewOutboxRepository(db)
	publisher := events.NewOutboxPublisher(outboxRepo)
	auth := &fakeTokenParser{}

	registerUC := common.NewTransactionalUseCase(uow, register.New(usersRepo, identitiesRepo, refreshRepo, stubHasher{}, auth, publisher, time.Minute, time.Hour))
	loginUC := common.NewTransactionalUseCase(uow, login.New(usersRepo, identitiesRepo, refreshRepo, stubHasher{}, auth, time.Minute, time.Hour))
	refreshUC := common.NewTransactionalUseCase(uow, refresh.New(refreshRepo, auth, time.Minute, time.Hour))
	meUC := common.NewTransactionalUseCase(uow, profile.NewGet(usersRepo))
	profileUC := common.NewTransactionalUseCase(uow, profile.NewUpdate(usersRepo))
	linkUC := common.NewTransactionalUseCase(uow, link.New(identitiesRepo))

	svc := usersapp.NewService(
		common.UseCaseHandler(registerUC),
		common.UseCaseHandler(loginUC),
		common.UseCaseHandler(refreshUC),
		common.UseCaseHandler(meUC),
		common.UseCaseHandler(profileUC),
		common.UseCaseHandler(linkUC),
	)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT\s+id::text,\s+user_id::text,\s+provider,\s+provider_user_id,\s+COALESCE\(secret_hash, ''\),\s+created_at\s+FROM auth_identities\s+WHERE provider = \$1 AND provider_user_id = \$2\s+LIMIT 1`).
		WithArgs("email", "user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_user_id", "secret_hash", "created_at"}))
	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO auth_identities`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "email", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO auth_refresh_tokens`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO user_events_outbox`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	server := newTestServer(svc, auth)
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "password123", "display_name": "User"})
	resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("db expectations: %v", err)
	}
}

func TestRegisterEndpointRollbackOnFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	uow := pdb.NewUnitOfWork(db)
	usersRepo := usersdb.NewUserRepo(db)
	identitiesRepo := usersdb.NewIdentityRepo(db)
	refreshRepo := usersdb.NewRefreshRepo(db)
	auth := &fakeTokenParser{}

	registerUC := common.NewTransactionalUseCase(uow, register.New(usersRepo, identitiesRepo, refreshRepo, stubHasher{}, auth, events.NewOutboxPublisher(events.NewOutboxRepository(db)), time.Minute, time.Hour))
	loginUC := common.NewTransactionalUseCase(uow, login.New(usersRepo, identitiesRepo, refreshRepo, stubHasher{}, auth, time.Minute, time.Hour))
	refreshUC := common.NewTransactionalUseCase(uow, refresh.New(refreshRepo, auth, time.Minute, time.Hour))
	meUC := common.NewTransactionalUseCase(uow, profile.NewGet(usersRepo))
	profileUC := common.NewTransactionalUseCase(uow, profile.NewUpdate(usersRepo))
	linkUC := common.NewTransactionalUseCase(uow, link.New(identitiesRepo))

	svc := usersapp.NewService(
		common.UseCaseHandler(registerUC),
		common.UseCaseHandler(loginUC),
		common.UseCaseHandler(refreshUC),
		common.UseCaseHandler(meUC),
		common.UseCaseHandler(profileUC),
		common.UseCaseHandler(linkUC),
	)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT\s+id::text,\s+user_id::text,\s+provider,\s+provider_user_id,\s+COALESCE\(secret_hash, ''\),\s+created_at\s+FROM auth_identities\s+WHERE provider = \$1 AND provider_user_id = \$2\s+LIMIT 1`).
		WithArgs("email", "user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_user_id", "secret_hash", "created_at"}))
	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO auth_identities`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "email", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO auth_refresh_tokens`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("fail"))
	mock.ExpectRollback()

	server := newTestServer(svc, auth)
	defer server.Close()

	body, _ := json.Marshal(map[string]string{"email": "user@example.com", "password": "password123", "display_name": "User"})
	resp, err := http.Post(server.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 on failure, got %d", resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("db expectations: %v", err)
	}
}
