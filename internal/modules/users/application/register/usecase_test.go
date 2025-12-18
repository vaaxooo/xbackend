package register

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type registerUnitOfWorkMock struct {
	called bool
	err    error
}

func (m *registerUnitOfWorkMock) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	m.called = true
	if m.err != nil {
		return m.err
	}
	return fn(ctx)
}

type registerUsersRepoMock struct {
	created []domain.User
	err     error
}

func (m *registerUsersRepoMock) Create(_ context.Context, u domain.User) error {
	m.created = append(m.created, u)
	return m.err
}

func (m *registerUsersRepoMock) GetByID(context.Context, domain.UserID) (domain.User, bool, error) {
	return domain.User{}, false, errors.New("not implemented")
}

func (m *registerUsersRepoMock) UpdateProfile(context.Context, domain.User) (domain.User, error) {
	return domain.User{}, errors.New("not implemented")
}

type registerIdentityRepoMock struct {
	created []domain.Identity
	exists  bool
	err     error
}

func (m *registerIdentityRepoMock) Create(_ context.Context, ident domain.Identity) error {
	m.created = append(m.created, ident)
	return m.err
}

func (m *registerIdentityRepoMock) GetByProvider(context.Context, string, string) (domain.Identity, bool, error) {
	return domain.Identity{}, m.exists, m.err
}

func (m *registerIdentityRepoMock) GetByUserAndProvider(context.Context, domain.UserID, string) (domain.Identity, bool, error) {
	return domain.Identity{}, false, errors.New("not implemented")
}

type registerRefreshRepoMock struct {
	created []domain.RefreshToken
	err     error
}

func (m *registerRefreshRepoMock) Create(_ context.Context, token domain.RefreshToken) error {
	m.created = append(m.created, token)
	return m.err
}

func (m *registerRefreshRepoMock) GetByHash(context.Context, string) (domain.RefreshToken, bool, error) {
	return domain.RefreshToken{}, false, errors.New("not implemented")
}

func (m *registerRefreshRepoMock) Revoke(context.Context, string) error { return nil }

type registerHasherMock struct {
	hash string
	err  error
}

func (m *registerHasherMock) Hash(context.Context, string) (string, error)  { return m.hash, m.err }
func (m *registerHasherMock) Compare(context.Context, string, string) error { return nil }

type registerTokenIssuerMock struct {
	token string
	err   error
}

func (m *registerTokenIssuerMock) Issue(_ string, _ time.Duration) (string, error) {
	return m.token, m.err
}

func TestRegisterSuccess(t *testing.T) {
	uow := &registerUnitOfWorkMock{}
	users := &registerUsersRepoMock{}
	identities := &registerIdentityRepoMock{}
	refreshRepo := &registerRefreshRepoMock{}
	hasher := &registerHasherMock{hash: "hashed"}
	issuer := &registerTokenIssuerMock{token: "access-token"}

	uc := New(uow, users, identities, refreshRepo, hasher, issuer, time.Minute, time.Hour)

	out, err := uc.Execute(context.Background(), Input{Email: "user@example.com", Password: "password123", DisplayName: "User"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !uow.called {
		t.Fatalf("expected unit of work to be called")
	}
	if len(users.created) != 1 || len(identities.created) != 1 || len(refreshRepo.created) != 1 {
		t.Fatalf("expected repositories to be invoked")
	}

	tokenHash := common.HashToken(out.RefreshToken)
	if refreshRepo.created[0].TokenHash != tokenHash {
		t.Fatalf("refresh token hash mismatch")
	}
	if refreshRepo.created[0].ExpiresAt.Sub(refreshRepo.created[0].CreatedAt) != time.Hour {
		t.Fatalf("unexpected refresh ttl")
	}
	if out.AccessToken != "access-token" {
		t.Fatalf("unexpected access token: %s", out.AccessToken)
	}
	if out.DisplayName != "User" || out.UserID == "" {
		t.Fatalf("expected user info in response: %+v", out)
	}
}

func TestRegisterEmailConflict(t *testing.T) {
	uc := New(&registerUnitOfWorkMock{}, &registerUsersRepoMock{}, &registerIdentityRepoMock{exists: true}, &registerRefreshRepoMock{}, &registerHasherMock{}, &registerTokenIssuerMock{}, 0, 0)

	_, err := uc.Execute(context.Background(), Input{Email: "user@example.com", Password: "password123"})
	if !errors.Is(err, domain.ErrEmailAlreadyUsed) {
		t.Fatalf("expected ErrEmailAlreadyUsed, got %v", err)
	}
}

func TestRegisterWeakPassword(t *testing.T) {
	uc := New(&registerUnitOfWorkMock{}, &registerUsersRepoMock{}, &registerIdentityRepoMock{}, &registerRefreshRepoMock{}, &registerHasherMock{}, &registerTokenIssuerMock{}, 0, 0)

	_, err := uc.Execute(context.Background(), Input{Email: "user@example.com", Password: "short"})
	if !errors.Is(err, domain.ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}
