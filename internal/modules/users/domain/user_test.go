package domain

import (
	"errors"
	"testing"
	"time"
)

func TestUserApplyPatch(t *testing.T) {
	now := time.Now()
	displayName, _ := NewDisplayName(" Name ")
	user := NewUser(UserID("id"), displayName, now)
	first := " John "
	last := "Doe"
	display := "New"
	avatar := " http://example.com "

	patched, err := user.ApplyPatch(ProfilePatch{
		FirstName:   &first,
		LastName:    &last,
		DisplayName: &display,
		AvatarURL:   &avatar,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !patched.ProfileCustomized {
		t.Fatalf("expected ProfileCustomized to be true")
	}
	if patched.FirstName != "John" || patched.LastName != "Doe" || patched.DisplayName != "New" {
		t.Fatalf("unexpected patched names: %+v", patched)
	}
	if patched.AvatarURL != "http://example.com" {
		t.Fatalf("unexpected avatar url: %s", patched.AvatarURL)
	}
}

func TestUserApplyPatch_InvalidDisplayName(t *testing.T) {
	now := time.Now()
	displayName, _ := NewDisplayName("Name")
	user := NewUser(UserID("id"), displayName, now)
	tooShort := " "

	if _, err := user.ApplyPatch(ProfilePatch{DisplayName: &tooShort}); !errors.Is(err, ErrInvalidDisplayName) {
		t.Fatalf("expected ErrInvalidDisplayName, got %v", err)
	}
}

func TestUserApplyPatch_InvalidAvatar(t *testing.T) {
	now := time.Now()
	displayName, _ := NewDisplayName("Name")
	user := NewUser(UserID("id"), displayName, now)
	invalid := "ftp://example.com/avatar.png"

	if _, err := user.ApplyPatch(ProfilePatch{AvatarURL: &invalid}); !errors.Is(err, ErrInvalidAvatarURL) {
		t.Fatalf("expected ErrInvalidAvatarURL, got %v", err)
	}
}
