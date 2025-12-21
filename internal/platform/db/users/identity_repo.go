package usersdb

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
	pdb "github.com/vaaxooo/xbackend/internal/platform/db"
)

type IdentityRepo struct {
	db *sql.DB
}

func NewIdentityRepo(db *sql.DB) *IdentityRepo {
	return &IdentityRepo{db: db}
}

func (r *IdentityRepo) Create(ctx context.Context, identity domain.Identity) error {
	const q = `
        INSERT INTO auth_identities (id, user_id, provider, provider_user_id, secret_hash, email_confirmed_at, totp_secret, totp_confirmed_at, created_at)
        VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9)
    `
	exec := pdb.Executor(ctx, r.db)
	_, err := exec.ExecContext(ctx, q,
		identity.ID,
		identity.UserID.String(),
		identity.Provider,
		identity.ProviderUserID,
		nullIfEmpty(identity.SecretHash.String()),
		identity.EmailVerifiedAt,
		nullIfEmpty(identity.TOTPSecret),
		identity.TOTPConfirmedAt,
		identity.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			if identity.Provider == "email" {
				return domain.ErrEmailAlreadyUsed
			}
			return domain.ErrIdentityAlreadyLinked
		}
		return err
	}
	return nil
}

func (r *IdentityRepo) GetByProvider(ctx context.Context, provider string, providerUserID string) (domain.Identity, bool, error) {
	const q = `
        SELECT
            id::text,
            user_id::text,
            provider,
            provider_user_id,
            COALESCE(secret_hash, ''),
            email_confirmed_at,
            COALESCE(totp_secret, ''),
            totp_confirmed_at,
            created_at
        FROM auth_identities
        WHERE provider = $1 AND provider_user_id = $2
        LIMIT 1
    `
	var i domain.Identity
	var userID string
	var secretHash string
	var confirmedAt sql.NullTime
	var totpConfirmed sql.NullTime
	err := pdb.Executor(ctx, r.db).QueryRowContext(ctx, q, provider, providerUserID).Scan(
		&i.ID, &userID, &i.Provider, &i.ProviderUserID, &secretHash, &confirmedAt, &i.TOTPSecret, &totpConfirmed, &i.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Identity{}, false, nil
	}
	if err != nil {
		return domain.Identity{}, false, err
	}
	i.UserID = domain.UserID(userID)
	i.SecretHash = domain.PasswordHash(secretHash)
	if confirmedAt.Valid {
		t := confirmedAt.Time
		i.EmailVerifiedAt = &t
	}
	if totpConfirmed.Valid {
		t := totpConfirmed.Time
		i.TOTPConfirmedAt = &t
	}
	return i, true, nil
}

func (r *IdentityRepo) GetByUserAndProvider(ctx context.Context, userID domain.UserID, provider string) (domain.Identity, bool, error) {
	const q = `
        SELECT
            id::text,
            user_id::text,
            provider,
            provider_user_id,
            COALESCE(secret_hash, ''),
            email_confirmed_at,
            COALESCE(totp_secret, ''),
            totp_confirmed_at,
            created_at
        FROM auth_identities
        WHERE user_id = $1::uuid AND provider = $2
        LIMIT 1
    `
	var i domain.Identity
	var id string
	var secretHash string
	var confirmedAt sql.NullTime
	var totpConfirmed sql.NullTime
	err := pdb.Executor(ctx, r.db).QueryRowContext(ctx, q, userID.String(), provider).Scan(
		&i.ID, &id, &i.Provider, &i.ProviderUserID, &secretHash, &confirmedAt, &i.TOTPSecret, &totpConfirmed, &i.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Identity{}, false, nil
	}
	if err != nil {
		return domain.Identity{}, false, err
	}
	i.UserID = domain.UserID(id)
	i.SecretHash = domain.PasswordHash(secretHash)
	if confirmedAt.Valid {
		t := confirmedAt.Time
		i.EmailVerifiedAt = &t
	}
	if totpConfirmed.Valid {
		t := totpConfirmed.Time
		i.TOTPConfirmedAt = &t
	}
	return i, true, nil
}

func (r *IdentityRepo) Update(ctx context.Context, identity domain.Identity) error {
	const q = `
        UPDATE auth_identities
        SET secret_hash = $2,
            email_confirmed_at = $3,
            totp_secret = $4,
            totp_confirmed_at = $5
        WHERE id = $1::uuid
    `
	_, err := pdb.Executor(ctx, r.db).ExecContext(ctx, q,
		identity.ID,
		nullIfEmpty(identity.SecretHash.String()),
		identity.EmailVerifiedAt,
		nullIfEmpty(identity.TOTPSecret),
		identity.TOTPConfirmedAt,
	)
	return err
}

func isUniqueViolation(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
