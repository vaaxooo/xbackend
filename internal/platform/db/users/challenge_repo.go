package usersdb

import (
	"context"
	"database/sql"
	"strings"

	"github.com/lib/pq"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
	pdb "github.com/vaaxooo/xbackend/internal/platform/db"
)

type ChallengeRepo struct {
	db *sql.DB
}

func NewChallengeRepo(db *sql.DB) *ChallengeRepo {
	return &ChallengeRepo{db: db}
}

func (r *ChallengeRepo) Create(ctx context.Context, challenge domain.Challenge) error {
	const q = `
        INSERT INTO auth_challenges (
            id, user_id, challenge_type, required_steps, completed_steps, status, expires_at, session_fingerprint,
            attempts_left, lock_until, created_at, updated_at
        ) VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `
	_, err := pdb.Executor(ctx, r.db).ExecContext(
		ctx,
		q,
		challenge.ID,
		challenge.UserID.String(),
		challenge.Type,
		pq.Array(challenge.RequiredSteps),
		pq.Array(challenge.CompletedSteps),
		string(challenge.Status),
		challenge.ExpiresAt,
		nullIfEmpty(challenge.SessionFingerprint),
		challenge.AttemptsLeft,
		challenge.LockUntil,
		challenge.CreatedAt,
		challenge.UpdatedAt,
	)
	return err
}

func (r *ChallengeRepo) Update(ctx context.Context, challenge domain.Challenge) error {
	const q = `
        UPDATE auth_challenges
        SET required_steps=$2, completed_steps=$3, status=$4, expires_at=$5, session_fingerprint=$6,
            attempts_left=$7, lock_until=$8, updated_at=$9
        WHERE id=$1::uuid
    `
	_, err := pdb.Executor(ctx, r.db).ExecContext(
		ctx,
		q,
		challenge.ID,
		pq.Array(challenge.RequiredSteps),
		pq.Array(challenge.CompletedSteps),
		string(challenge.Status),
		challenge.ExpiresAt,
		nullIfEmpty(challenge.SessionFingerprint),
		challenge.AttemptsLeft,
		challenge.LockUntil,
		challenge.UpdatedAt,
	)
	return err
}

func (r *ChallengeRepo) GetByID(ctx context.Context, id string) (domain.Challenge, bool, error) {
	const q = `
        SELECT id::text, user_id::text, challenge_type, required_steps, completed_steps, status, expires_at,
               COALESCE(session_fingerprint, ''), attempts_left, lock_until, created_at, updated_at
        FROM auth_challenges
        WHERE id=$1::uuid
        LIMIT 1
    `
	var c domain.Challenge
	var required, completed []string
	var status string
	err := pdb.Executor(ctx, r.db).QueryRowContext(ctx, q, id).Scan(
		&c.ID,
		&c.UserID,
		&c.Type,
		pq.Array(&required),
		pq.Array(&completed),
		&status,
		&c.ExpiresAt,
		&c.SessionFingerprint,
		&c.AttemptsLeft,
		&c.LockUntil,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.Challenge{}, false, nil
	}
	if err != nil {
		return domain.Challenge{}, false, err
	}
	c.RequiredSteps = toChallengeSteps(required)
	c.CompletedSteps = toChallengeSteps(completed)
	c.Status = domain.ChallengeStatus(status)
	if strings.TrimSpace(c.SessionFingerprint) == "" {
		c.SessionFingerprint = ""
	}
	return c, true, nil
}

func (r *ChallengeRepo) GetPendingByUser(ctx context.Context, userID domain.UserID) (domain.Challenge, bool, error) {
	const q = `
        SELECT id::text, user_id::text, challenge_type, required_steps, completed_steps, status, expires_at,
               COALESCE(session_fingerprint, ''), attempts_left, lock_until, created_at, updated_at
        FROM auth_challenges
        WHERE user_id=$1::uuid AND status='pending'
        ORDER BY created_at DESC
        LIMIT 1
    `
	var c domain.Challenge
	var required, completed []string
	var status string
	err := pdb.Executor(ctx, r.db).QueryRowContext(ctx, q, userID.String()).Scan(
		&c.ID,
		&c.UserID,
		&c.Type,
		pq.Array(&required),
		pq.Array(&completed),
		&status,
		&c.ExpiresAt,
		&c.SessionFingerprint,
		&c.AttemptsLeft,
		&c.LockUntil,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.Challenge{}, false, nil
	}
	if err != nil {
		return domain.Challenge{}, false, err
	}
	c.RequiredSteps = toChallengeSteps(required)
	c.CompletedSteps = toChallengeSteps(completed)
	c.Status = domain.ChallengeStatus(status)
	if strings.TrimSpace(c.SessionFingerprint) == "" {
		c.SessionFingerprint = ""
	}
	return c, true, nil
}

func toChallengeSteps(values []string) []domain.ChallengeStep {
	steps := make([]domain.ChallengeStep, 0, len(values))
	for _, v := range values {
		steps = append(steps, domain.ChallengeStep(v))
	}
	return steps
}

var _ domain.ChallengeRepository = (*ChallengeRepo)(nil)
