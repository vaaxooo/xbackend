package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type profileUsersRepoMock struct {
	user    domain.User
	updated domain.User
	err     error
}

func (m *profileUsersRepoMock) Create(context.Context, domain.User) error {
	return errors.New("not implemented")
}

func (m *profileUsersRepoMock) GetByID(context.Context, domain.UserID) (domain.User, bool, error) {
	if m.err != nil {
		return domain.User{}, false, m.err
	}
	if m.user.ID == "" {
		return domain.User{}, false, nil
	}
	return m.user, true, nil
}

func (m *profileUsersRepoMock) UpdateProfile(_ context.Context, in domain.User) (domain.User, error) {
	if m.err != nil {
		return domain.User{}, m.err
	}
	m.updated = in
	return in, nil
}

func TestGetProfile(t *testing.T) {
	repo := &profileUsersRepoMock{user: domain.User{ID: "user", DisplayName: "User"}}
	uc := NewGet(repo)

	out, err := uc.Execute(context.Background(), GetInput{UserID: "user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.UserID != "user" || out.DisplayName != "User" {
		t.Fatalf("unexpected profile: %+v", out)
	}
}

func TestGetProfileUnauthorized(t *testing.T) {
	repo := &profileUsersRepoMock{}
	uc := NewGet(repo)

	if _, err := uc.Execute(context.Background(), GetInput{UserID: "   "}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestUpdateProfile(t *testing.T) {
	user := domain.User{ID: "user", DisplayName: "Old"}
	repo := &profileUsersRepoMock{user: user}
	uc := NewUpdate(repo)

	first := "John"
	out, err := uc.Execute(context.Background(), UpdateInput{UserID: "user", FirstName: &first})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.FirstName != "John" {
		t.Fatalf("expected first name updated, got %+v", out)
	}
	if repo.updated.FirstName != "John" || !repo.updated.ProfileCustomized {
		t.Fatalf("expected patched user saved")
	}
}

func TestUpdateProfile_InvalidAvatar(t *testing.T) {
	user := domain.User{ID: "user", DisplayName: "Old"}
	repo := &profileUsersRepoMock{user: user}
	uc := NewUpdate(repo)

	avatar := "ftp://example.com/avatar.png"
	if _, err := uc.Execute(context.Background(), UpdateInput{UserID: "user", AvatarURL: &avatar}); !errors.Is(err, domain.ErrInvalidAvatarURL) {
		t.Fatalf("expected invalid avatar error, got %v", err)
	}
}
