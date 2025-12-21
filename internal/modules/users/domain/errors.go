package domain

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrWeakPassword       = errors.New("weak password")
	ErrEmailAlreadyUsed   = errors.New("email already used")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrTwoFactorRequired  = errors.New("two factor required")
	ErrInvalidTwoFactor   = errors.New("invalid two factor code")
	ErrTooManyRequests    = errors.New("too many requests")
	ErrInvalidDisplayName = errors.New("invalid display name")
	ErrInvalidAvatarURL   = errors.New("invalid avatar url")

	ErrUnauthorized          = errors.New("unauthorized")
	ErrIdentityAlreadyLinked = errors.New("identity already linked")
	ErrRefreshTokenInvalid   = errors.New("refresh token invalid")
)
