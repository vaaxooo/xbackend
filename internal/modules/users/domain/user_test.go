package domain

import (
	"testing"
	"time"
)

func TestUserApplyPatch(t *testing.T) {
	now := time.Now()
	user := NewUser(UserID("id"), " Name ", now)
	first := " John "
	last := "Doe"
	display := "New"
	avatar := " http://example.com "

	patched := user.ApplyPatch(ProfilePatch{
		FirstName:   &first,
		LastName:    &last,
		DisplayName: &display,
		AvatarURL:   &avatar,
	})

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
