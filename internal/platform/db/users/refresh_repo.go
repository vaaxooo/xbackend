package usersdb

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
	pdb "github.com/vaaxooo/xbackend/internal/platform/db"
)

type RefreshRepo struct {
	db         *sql.DB
	cleanupTTL time.Duration
}

func NewRefreshRepo(db *sql.DB, cleanupTTL time.Duration) *RefreshRepo {
	if cleanupTTL < 0 {
		cleanupTTL = 0
	} else if cleanupTTL == 0 {
		cleanupTTL = 90 * 24 * time.Hour
	}
	return &RefreshRepo{db: db, cleanupTTL: cleanupTTL}
}

func (r *RefreshRepo) Create(ctx context.Context, t domain.RefreshToken) error {
	r.cleanupStale(ctx, time.Now().UTC())
	const q = `
        INSERT INTO auth_refresh_tokens (id, user_id, token_hash, expires_at, revoked_at, created_at, user_agent, ip)
        VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8)
    `
	exec := pdb.Executor(ctx, r.db)
	_, err := exec.ExecContext(ctx, q,
		t.ID,
		t.UserID.String(),
		t.TokenHash,
		t.ExpiresAt,
		t.RevokedAt,
		t.CreatedAt,
		nullIfEmpty(t.UserAgent),
		nullIfEmpty(t.IP),
	)
	return err
}

func (r *RefreshRepo) Update(ctx context.Context, t domain.RefreshToken) error {
	r.cleanupStale(ctx, time.Now().UTC())
	const q = `
        UPDATE auth_refresh_tokens
        SET token_hash = $2,
            expires_at = $3,
            revoked_at = $4,
            created_at = $5,
            user_agent = $6,
            ip = $7
        WHERE id = $1::uuid
    `
	exec := pdb.Executor(ctx, r.db)
	_, err := exec.ExecContext(ctx, q,
		t.ID,
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
	row := pdb.Executor(ctx, r.db).QueryRowContext(ctx, q, tokenHash)
	t, err := scanRefreshToken(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RefreshToken{}, false, nil
	}
	if err != nil {
		return domain.RefreshToken{}, false, err
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

func (r *RefreshRepo) GetByID(ctx context.Context, tokenID string) (domain.RefreshToken, bool, error) {
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
        WHERE id = $1::uuid
        LIMIT 1
    `
	row := pdb.Executor(ctx, r.db).QueryRowContext(ctx, q, tokenID)
	t, err := scanRefreshToken(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RefreshToken{}, false, nil
	}
	if err != nil {
		return domain.RefreshToken{}, false, err
	}
	return t, true, nil
}

func (r *RefreshRepo) ListByUser(ctx context.Context, userID domain.UserID) ([]domain.RefreshToken, error) {
	r.cleanupStale(ctx, time.Now().UTC())
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
        WHERE user_id = $1::uuid AND revoked_at IS NULL AND expires_at > $2
        ORDER BY created_at DESC
        LIMIT 15
    `
	rows, err := pdb.Executor(ctx, r.db).QueryContext(ctx, q, userID.String(), time.Now().UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens := make([]domain.RefreshToken, 0)
	for rows.Next() {
		t, scanErr := scanRefreshToken(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		tokens = append(tokens, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *RefreshRepo) FindActiveByFingerprint(ctx context.Context, userID domain.UserID, userAgent, ip string, now time.Time) (domain.RefreshToken, bool, error) {
	r.cleanupStale(ctx, now)
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
        WHERE user_id = $1::uuid
          AND revoked_at IS NULL
          AND expires_at > $2
          AND COALESCE(user_agent, '') = $3
          AND COALESCE(ip, '') = $4
        ORDER BY created_at DESC
        LIMIT 1
    `
	row := pdb.Executor(ctx, r.db).QueryRowContext(ctx, q, userID.String(), now, userAgent, ip)
	t, err := scanRefreshToken(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RefreshToken{}, false, nil
	}
	if err != nil {
		return domain.RefreshToken{}, false, err
	}
	return t, true, nil
}

func (r *RefreshRepo) RevokeAllExcept(ctx context.Context, userID domain.UserID, keepIDs []string) error {
	ids := keepIDs
	if ids == nil {
		ids = []string{}
	}

	now := time.Now().UTC()
	const q = `
        UPDATE auth_refresh_tokens
        SET revoked_at = $3
        WHERE user_id = $1::uuid AND revoked_at IS NULL AND id <> ALL($2::uuid[])
    `
	_, err := pdb.Executor(ctx, r.db).ExecContext(ctx, q, userID.String(), pq.Array(ids), now)
	return err
}

func (r *RefreshRepo) cleanupStale(ctx context.Context, now time.Time) {
	if r.cleanupTTL <= 0 {
		return
	}
	cutoff := now.Add(-r.cleanupTTL)
	const q = `
        DELETE FROM auth_refresh_tokens
        WHERE (revoked_at IS NOT NULL AND revoked_at < $1)
           OR (expires_at < $1)
    `
	_, _ = pdb.Executor(ctx, r.db).ExecContext(ctx, q, cutoff)
}

type refreshScanner interface {
	Scan(dest ...any) error
}

func scanRefreshToken(scanner refreshScanner) (domain.RefreshToken, error) {
	var t domain.RefreshToken
	var revokedAt sql.NullTime
	var userID string

	err := scanner.Scan(
		&t.ID,
		&userID,
		&t.TokenHash,
		&t.ExpiresAt,
		&revokedAt,
		&t.CreatedAt,
		&t.UserAgent,
		&t.IP,
	)
	if err != nil {
		return domain.RefreshToken{}, err
	}
	if revokedAt.Valid {
		v := revokedAt.Time
		t.RevokedAt = &v
	}
	t.UserID = domain.UserID(userID)
	return t, nil
}
