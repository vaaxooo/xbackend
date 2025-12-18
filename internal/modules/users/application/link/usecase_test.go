package link

import (
	"context"
	"errors"
	"testing"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type linkIdentityRepoMock struct {
	created   []domain.Identity
	available bool
	err       error
}

func (m *linkIdentityRepoMock) Create(_ context.Context, i domain.Identity) error {
	m.created = append(m.created, i)
	return m.err
}

func (m *linkIdentityRepoMock) GetByProvider(context.Context, string, string) (domain.Identity, bool, error) {
	return domain.Identity{}, !m.available, m.err
}

func (m *linkIdentityRepoMock) GetByUserAndProvider(context.Context, domain.UserID, string) (domain.Identity, bool, error) {
	return domain.Identity{}, !m.available, m.err
}

func TestLinkProvider(t *testing.T) {
	repo := &linkIdentityRepoMock{available: true}
	uc := New(repo)

	out, err := uc.Execute(context.Background(), Input{UserID: "user", Provider: "github", ProviderUserID: "gh-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Linked || len(repo.created) != 1 {
		t.Fatalf("expected identity to be linked")
	}
}

func TestLinkProviderUnavailable(t *testing.T) {
	repo := &linkIdentityRepoMock{}
	uc := New(repo)

	_, err := uc.Execute(context.Background(), Input{UserID: "user", Provider: "github", ProviderUserID: "gh-1"})
	if !errors.Is(err, domain.ErrIdentityAlreadyLinked) {
		t.Fatalf("expected conflict error, got %v", err)
	}
}
