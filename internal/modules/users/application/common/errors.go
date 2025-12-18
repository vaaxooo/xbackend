package common

import (
	"errors"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

var ErrInternal = errors.New("internal error")

// NormalizeError passes domain errors through unchanged and wraps
// infrastructure/unknown errors into a consistent application-level error.
func NormalizeError(err error) error {
	if err == nil {
		return nil
	}
	if isDomainError(err) {
		return err
	}
	return errors.Join(ErrInternal, err)
}

func isDomainError(err error) bool {
	switch {
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrWeakPassword),
		errors.Is(err, domain.ErrEmailAlreadyUsed),
		errors.Is(err, domain.ErrInvalidCredentials),
		errors.Is(err, domain.ErrUnauthorized),
		errors.Is(err, domain.ErrIdentityAlreadyLinked),
		errors.Is(err, domain.ErrRefreshTokenInvalid):
		return true
	default:
		return false
	}
}
