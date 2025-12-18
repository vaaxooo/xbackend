package refresh

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type refreshUnitOfWorkMock struct{ called bool }

func (m *refreshUnitOfWorkMock) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	m.called = true
	return fn(ctx)
}

type refreshRepoMock struct {
	stored  domain.RefreshToken
	found   bool
	err     error
	revoked string
	created []domain.RefreshToken
}

func (m *refreshRepoMock) Create(_ context.Context, t domain.RefreshToken) error {
	m.created = append(m.created, t)
	return m.err
}

func (m *refreshRepoMock) GetByHash(_ context.Context, hash string) (domain.RefreshToken, bool, error) {
	return m.stored, m.found, m.err
}

func (m *refreshRepoMock) Revoke(_ context.Context, id string) error {
	m.revoked = id
	return m.err
}

type refreshIssuerMock struct{ token string }

func (m *refreshIssuerMock) Issue(_ string, _ time.Duration) (string, error) { return m.token, nil }

func TestRefreshSuccess(t *testing.T) {
	now := time.Now().UTC()
	repo := &refreshRepoMock{stored: domain.RefreshToken{ID: "id", UserID: "user", TokenHash: common.HashToken("old"), ExpiresAt: now.Add(time.Hour)}, found: true}
	uow := &refreshUnitOfWorkMock{}
	uc := common.NewTransactionalUseCase(uow, New(repo, &refreshIssuerMock{token: "access"}, time.Minute, time.Hour))

	out, err := uc.Execute(context.Background(), Input{RefreshToken: "old"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.AccessToken != "access" || out.RefreshToken == "" {
		t.Fatalf("unexpected output: %+v", out)
	}
	if !uow.called || repo.revoked != "id" || len(repo.created) != 1 {
		t.Fatalf("expected token rotation inside transaction")
	}
}

func TestRefreshInvalid(t *testing.T) {
	repo := &refreshRepoMock{found: false}
	uc := common.NewTransactionalUseCase(&refreshUnitOfWorkMock{}, New(repo, &refreshIssuerMock{}, 0, 0))

	if _, err := uc.Execute(context.Background(), Input{RefreshToken: ""}); !errors.Is(err, domain.ErrRefreshTokenInvalid) {
		t.Fatalf("expected invalid token on empty input, got %v", err)
	}

	repo = &refreshRepoMock{stored: domain.RefreshToken{ID: "id", ExpiresAt: time.Now().Add(-time.Hour)}, found: true}
	uc = common.NewTransactionalUseCase(&refreshUnitOfWorkMock{}, New(repo, &refreshIssuerMock{}, 0, 0))
	if _, err := uc.Execute(context.Background(), Input{RefreshToken: "expired"}); !errors.Is(err, domain.ErrRefreshTokenInvalid) {
		t.Fatalf("expected invalid token on expired, got %v", err)
	}
	if repo.revoked != "id" {
		t.Fatalf("expected expired token to be revoked")
	}
}
