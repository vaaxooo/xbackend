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
        INSERT INTO auth_identities (id, user_id, provider, provider_user_id, secret_hash, created_at)
        VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6)
    `
	exec := pdb.Executor(ctx, r.db)
	_, err := exec.ExecContext(ctx, q,
		identity.ID,
		identity.UserID.String(),
		identity.Provider,
		identity.ProviderUserID,
		nullIfEmpty(identity.SecretHash.String()),
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
            created_at
        FROM auth_identities
        WHERE provider = $1 AND provider_user_id = $2
        LIMIT 1
    `
	var i domain.Identity
	var userID string
	var secretHash string
	err := pdb.Executor(ctx, r.db).QueryRowContext(ctx, q, provider, providerUserID).Scan(
		&i.ID, &userID, &i.Provider, &i.ProviderUserID, &secretHash, &i.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Identity{}, false, nil
	}
	if err != nil {
		return domain.Identity{}, false, err
	}
	i.UserID = domain.UserID(userID)
	i.SecretHash = domain.PasswordHash(secretHash)
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
            created_at
        FROM auth_identities
        WHERE user_id = $1::uuid AND provider = $2
        LIMIT 1
    `
	var i domain.Identity
	var id string
	var secretHash string
	err := pdb.Executor(ctx, r.db).QueryRowContext(ctx, q, userID.String(), provider).Scan(
		&i.ID, &id, &i.Provider, &i.ProviderUserID, &secretHash, &i.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Identity{}, false, nil
	}
	if err != nil {
		return domain.Identity{}, false, err
	}
	i.UserID = domain.UserID(id)
	i.SecretHash = domain.PasswordHash(secretHash)
	return i, true, nil
}

func isUniqueViolation(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
