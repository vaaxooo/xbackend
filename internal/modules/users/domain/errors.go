package domain

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrWeakPassword       = errors.New("weak password")
	ErrEmailAlreadyUsed   = errors.New("email already used")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidDisplayName = errors.New("invalid display name")
	ErrInvalidAvatarURL   = errors.New("invalid avatar url")

	ErrUnauthorized          = errors.New("unauthorized")
	ErrIdentityAlreadyLinked = errors.New("identity already linked")
	ErrRefreshTokenInvalid   = errors.New("refresh token invalid")
)
