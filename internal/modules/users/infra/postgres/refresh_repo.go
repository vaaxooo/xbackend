package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
	pdb "github.com/vaaxooo/xbackend/internal/platform/db"
)

type RefreshRepo struct {
	db *sql.DB
}

func NewRefreshRepo(db *sql.DB) *RefreshRepo {
	return &RefreshRepo{db: db}
}

func (r *RefreshRepo) Create(ctx context.Context, t domain.RefreshToken) error {
	const q = `
        INSERT INTO auth_refresh_tokens (id, user_id, token_hash, expires_at, revoked_at, created_at, user_agent, ip)
        VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8)
    `
	exec := pdb.Executor(ctx, r.db)
	_, err := exec.ExecContext(ctx, q,
		t.ID,
		t.UserID,
		t.TokenHash,
		t.ExpiresAt,
		t.RevokedAt,
		t.CreatedAt,
		nullIfEmpty(t.UserAgent),
		nullIfEmpty(t.IP),
	)
	return err
}

func (r *RefreshRepo) GetByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, bool, error) {
	const q = `
        SELECT
            id::text,
            user_id::text,
            token_hash,
            expires_at,
            revoked_at,
            created_at,
            COALESCE(user_agent, ''),
            COALESCE(ip, '')
        FROM auth_refresh_tokens
        WHERE token_hash = $1
        LIMIT 1
    `
	var t domain.RefreshToken
	var revokedAt sql.NullTime

	err := pdb.Executor(ctx, r.db).QueryRowContext(ctx, q, tokenHash).Scan(
		&t.ID,
		&t.UserID,
		&t.TokenHash,
		&t.ExpiresAt,
		&revokedAt,
		&t.CreatedAt,
		&t.UserAgent,
		&t.IP,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RefreshToken{}, false, nil
	}
	if err != nil {
		return domain.RefreshToken{}, false, err
	}
	if revokedAt.Valid {
		v := revokedAt.Time
		t.RevokedAt = &v
	}
	return t, true, nil
}

func (r *RefreshRepo) Revoke(ctx context.Context, tokenID string) error {
	now := time.Now().UTC()
	const q = `
        UPDATE auth_refresh_tokens
        SET revoked_at = $2
        WHERE id = $1::uuid AND revoked_at IS NULL
    `
	_, err := pdb.Executor(ctx, r.db).ExecContext(ctx, q, tokenID, now)
	return err
}
