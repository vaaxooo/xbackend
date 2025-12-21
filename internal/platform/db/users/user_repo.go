package usersdb

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
	pdb "github.com/vaaxooo/xbackend/internal/platform/db"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user domain.User) error {
	const q = `
INSERT INTO users (id, display_name, avatar_url, profile_customized, suspended, suspension_reason, blocked_until, created_at)
VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8)
`
	exec := pdb.Executor(ctx, r.db)
	_, err := exec.ExecContext(
		ctx,
		q,
		user.ID.String(),
		nullIfEmpty(user.DisplayName),
		nullIfEmpty(user.AvatarURL),
		user.ProfileCustomized,
		user.Suspended,
		nullIfEmpty(user.SuspensionReason),
		user.BlockedUntil,
		user.CreatedAt,
	)
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, userID domain.UserID) (domain.User, bool, error) {
	const q = `
                SELECT
                        id::text,
			COALESCE(first_name, ''),
			COALESCE(last_name, ''),
			COALESCE(middle_name, ''),
COALESCE(display_name, ''),
COALESCE(avatar_url, ''),
profile_customized,
suspended,
COALESCE(suspension_reason, ''),
blocked_until,
created_at
FROM users
WHERE id = $1::uuid
LIMIT 1
`

	var u domain.User
	var id string
	err := pdb.Executor(ctx, r.db).QueryRowContext(ctx, q, userID.String()).Scan(
		&id,
		&u.FirstName,
		&u.LastName,
		&u.MiddleName,
		&u.DisplayName,
		&u.AvatarURL,
		&u.ProfileCustomized,
		&u.Suspended,
		&u.SuspensionReason,
		&u.BlockedUntil,
		&u.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, false, nil
	}
	if err != nil {
		return domain.User{}, false, err
	}
	u.ID = domain.UserID(id)
	return u, true, nil
}

func (r *UserRepo) UpdateProfile(
	ctx context.Context,
	in domain.User,
) (domain.User, error) {
	const q = `
        UPDATE users
        SET
            first_name = $2,
            last_name = $3,
            middle_name = $4,
            display_name = $5,
            avatar_url = $6,
            profile_customized = TRUE
        WHERE id = $1::uuid
RETURNING
id::text,
COALESCE(first_name, ''),
COALESCE(last_name, ''),
COALESCE(middle_name, ''),
COALESCE(display_name, ''),
COALESCE(avatar_url, ''),
profile_customized,
suspended,
COALESCE(suspension_reason, ''),
blocked_until,
created_at
`

	var u domain.User
	err := pdb.Executor(ctx, r.db).QueryRowContext(
		ctx,
		q,
		in.ID.String(),
		nullIfEmpty(in.FirstName),
		nullIfEmpty(in.LastName),
		nullIfEmpty(in.MiddleName),
		nullIfEmpty(in.DisplayName),
		nullIfEmpty(in.AvatarURL),
	).Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.MiddleName,
		&u.DisplayName,
		&u.AvatarURL,
		&u.ProfileCustomized,
		&u.Suspended,
		&u.SuspensionReason,
		&u.BlockedUntil,
		&u.CreatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}
	return u, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
