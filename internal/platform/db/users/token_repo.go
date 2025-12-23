package usersdb

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
	pdb "github.com/vaaxooo/xbackend/internal/platform/db"
)

type VerificationTokenRepo struct {
	db *sql.DB
}

func NewVerificationTokenRepo(db *sql.DB) *VerificationTokenRepo {
	return &VerificationTokenRepo{db: db}
}

func (r *VerificationTokenRepo) Create(ctx context.Context, token domain.VerificationToken) error {
	const q = `
        INSERT INTO auth_verification_tokens (id, identity_id, token_type, token_code, expires_at, used_at, created_at)
        VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7)
    `
	_, err := pdb.Executor(ctx, r.db).ExecContext(ctx, q,
		token.ID,
		token.IdentityID,
		string(token.Type),
		token.Code,
		token.ExpiresAt,
		token.UsedAt,
		token.CreatedAt,
	)
	return err
}

func (r *VerificationTokenRepo) GetLatest(ctx context.Context, identityID string, tokenType domain.TokenType) (domain.VerificationToken, bool, error) {
	const q = `
        SELECT id::text, identity_id::text, token_type, token_code, expires_at, used_at, created_at
        FROM auth_verification_tokens
        WHERE identity_id = $1::uuid AND token_type = $2
        ORDER BY created_at DESC
        LIMIT 1
    `
	return r.fetch(ctx, q, identityID, string(tokenType))
}

func (r *VerificationTokenRepo) GetByID(ctx context.Context, tokenID string) (domain.VerificationToken, bool, error) {
	const q = `
        SELECT id::text, identity_id::text, token_type, token_code, expires_at, used_at, created_at
        FROM auth_verification_tokens
        WHERE id = $1::uuid
    `
	return r.fetch(ctx, q, tokenID)
}

func (r *VerificationTokenRepo) GetByCode(ctx context.Context, identityID string, tokenType domain.TokenType, code string) (domain.VerificationToken, bool, error) {
	const q = `
        SELECT id::text, identity_id::text, token_type, token_code, expires_at, used_at, created_at
        FROM auth_verification_tokens
        WHERE identity_id = $1::uuid AND token_type = $2 AND token_code = $3
        ORDER BY created_at DESC
        LIMIT 1
    `
	return r.fetch(ctx, q, identityID, string(tokenType), code)
}

func (r *VerificationTokenRepo) fetch(ctx context.Context, query string, args ...any) (domain.VerificationToken, bool, error) {
	var t domain.VerificationToken
	var usedAt sql.NullTime
	err := pdb.Executor(ctx, r.db).QueryRowContext(ctx, query, args...).Scan(
		&t.ID,
		&t.IdentityID,
		&t.Type,
		&t.Code,
		&t.ExpiresAt,
		&usedAt,
		&t.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.VerificationToken{}, false, nil
	}
	if err != nil {
		return domain.VerificationToken{}, false, err
	}
	if usedAt.Valid {
		v := usedAt.Time
		t.UsedAt = &v
	}
	return t, true, nil
}

func (r *VerificationTokenRepo) MarkUsed(ctx context.Context, tokenID string, usedAt time.Time) error {
	const q = `
        UPDATE auth_verification_tokens
        SET used_at = $2
        WHERE id = $1::uuid
    `
	_, err := pdb.Executor(ctx, r.db).ExecContext(ctx, q, tokenID, usedAt)
	return err
}

var _ domain.VerificationTokenRepository = (*VerificationTokenRepo)(nil)
