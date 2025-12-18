package domain

import (
	"strings"
	"time"
)

type User struct {
	ID                UserID
	FirstName         string
	LastName          string
	MiddleName        string
	DisplayName       string
	AvatarURL         string
	ProfileCustomized bool
	CreatedAt         time.Time
}

func NewUser(id UserID, displayName DisplayName, createdAt time.Time) User {
	return User{
		ID:                id,
		DisplayName:       displayName.String(),
		AvatarURL:         "",
		ProfileCustomized: false,
		CreatedAt:         createdAt,
	}
}

type ProfilePatch struct {
	FirstName   *string
	LastName    *string
	MiddleName  *string
	DisplayName *string
	AvatarURL   *string
}

func (u User) ApplyPatch(p ProfilePatch) (User, error) {
	if p.FirstName != nil {
		u.FirstName = strings.TrimSpace(*p.FirstName)
	}
	if p.LastName != nil {
		u.LastName = strings.TrimSpace(*p.LastName)
	}
	if p.MiddleName != nil {
		u.MiddleName = strings.TrimSpace(*p.MiddleName)
	}
	if p.DisplayName != nil {
		displayName, err := NewDisplayName(*p.DisplayName)
		if err != nil {
			return User{}, err
		}
		u.DisplayName = displayName.String()
	}
	if p.AvatarURL != nil {
		avatarURL, err := NewAvatarURL(*p.AvatarURL)
		if err != nil {
			return User{}, err
		}
		u.AvatarURL = avatarURL.String()
	}
	u.ProfileCustomized = true
	return u, nil
}
