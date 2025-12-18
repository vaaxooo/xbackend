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

func NewUser(id UserID, displayName string, createdAt time.Time) User {
	return User{
		ID:                id,
		DisplayName:       strings.TrimSpace(displayName),
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

func (u User) ApplyPatch(p ProfilePatch) User {
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
		u.DisplayName = strings.TrimSpace(*p.DisplayName)
	}
	if p.AvatarURL != nil {
		u.AvatarURL = strings.TrimSpace(*p.AvatarURL)
	}
	u.ProfileCustomized = true
	return u
}
