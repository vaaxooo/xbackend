package login

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type loginUnitOfWorkMock struct {
	called bool
}

func (m *loginUnitOfWorkMock) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	m.called = true
	return fn(ctx)
}

type loginUsersRepoMock struct {
	user domain.User
	err  error
}

func (m *loginUsersRepoMock) Create(context.Context, domain.User) error {
	return errors.New("not implemented")
}
func (m *loginUsersRepoMock) GetByID(context.Context, domain.UserID) (domain.User, bool, error) {
	if m.err != nil {
		return domain.User{}, false, m.err
	}
	if m.user.ID == "" {
		return domain.User{}, false, nil
	}
	return m.user, true, nil
}
func (m *loginUsersRepoMock) UpdateProfile(context.Context, domain.User) (domain.User, error) {
	return domain.User{}, errors.New("not implemented")
}

type loginIdentityRepoMock struct {
	identity domain.Identity
	found    bool
	err      error
}

func (m *loginIdentityRepoMock) Create(context.Context, domain.Identity) error {
	return errors.New("not implemented")
}
func (m *loginIdentityRepoMock) GetByProvider(context.Context, string, string) (domain.Identity, bool, error) {
	return m.identity, m.found, m.err
}
func (m *loginIdentityRepoMock) GetByUserAndProvider(context.Context, domain.UserID, string) (domain.Identity, bool, error) {
	return domain.Identity{}, false, errors.New("not implemented")
}

type loginRefreshRepoMock struct{ created []domain.RefreshToken }

func (m *loginRefreshRepoMock) Create(_ context.Context, token domain.RefreshToken) error {
	m.created = append(m.created, token)
	return nil
}
func (m *loginRefreshRepoMock) GetByHash(context.Context, string) (domain.RefreshToken, bool, error) {
	return domain.RefreshToken{}, false, errors.New("not implemented")
}
func (m *loginRefreshRepoMock) Revoke(context.Context, string) error { return nil }

type loginHasherMock struct{ compareErr error }

func (m *loginHasherMock) Hash(context.Context, string) (string, error) {
	return "", errors.New("not implemented")
}
func (m *loginHasherMock) Compare(context.Context, string, string) error { return m.compareErr }

type loginIssuerMock struct{ token string }

func (m *loginIssuerMock) Issue(_ string, _ time.Duration) (string, error) { return m.token, nil }

func TestLoginSuccess(t *testing.T) {
	user := domain.User{ID: "user-1", DisplayName: "User"}
	identity := domain.Identity{UserID: user.ID, SecretHash: "hash"}

	uow := &loginUnitOfWorkMock{}
	uc := common.NewTransactionalUseCase(uow, New(&loginUsersRepoMock{user: user}, &loginIdentityRepoMock{identity: identity, found: true}, &loginRefreshRepoMock{}, &loginHasherMock{}, &loginIssuerMock{token: "access"}, time.Minute, time.Hour))

	out, err := uc.Execute(context.Background(), Input{Email: "user@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.UserID != user.ID.String() || out.AccessToken != "access" {
		t.Fatalf("unexpected output: %+v", out)
	}
	if !uow.called {
		t.Fatalf("expected refresh token created in transaction")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	uc := common.NewTransactionalUseCase(&loginUnitOfWorkMock{}, New(&loginUsersRepoMock{}, &loginIdentityRepoMock{}, &loginRefreshRepoMock{}, &loginHasherMock{}, &loginIssuerMock{}, 0, 0))

	if _, err := uc.Execute(context.Background(), Input{Email: "bad", Password: "pw"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials for bad email, got %v", err)
	}

	uc = common.NewTransactionalUseCase(&loginUnitOfWorkMock{}, New(&loginUsersRepoMock{}, &loginIdentityRepoMock{found: true, identity: domain.Identity{UserID: "user"}}, &loginRefreshRepoMock{}, &loginHasherMock{compareErr: errors.New("fail")}, &loginIssuerMock{}, 0, 0))
	if _, err := uc.Execute(context.Background(), Input{Email: "user@example.com", Password: "pw"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials for compare failure, got %v", err)
	}
}
