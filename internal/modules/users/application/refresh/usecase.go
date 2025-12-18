package refresh

import (
	"context"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type UseCase struct {
	uow         common.UnitOfWork
	refreshRepo domain.RefreshTokenRepository

	access     common.AccessTokenIssuer
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(
	uow common.UnitOfWork,
	refreshRepo domain.RefreshTokenRepository,
	access common.AccessTokenIssuer,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *UseCase {
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL == 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	return &UseCase{
		uow:         uow,
		refreshRepo: refreshRepo,
		access:      access,
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
	}
}

func (uc *UseCase) Execute(ctx context.Context, in Input) (Output, error) {
	if in.RefreshToken == "" {
		return Output{}, domain.ErrRefreshTokenInvalid
	}

	hash := common.HashToken(in.RefreshToken)
	stored, found, err := uc.refreshRepo.GetByHash(ctx, hash)
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}
	now := time.Now().UTC()
	if !found || !stored.IsValid(now) {
		if found && stored.RevokedAt == nil && now.After(stored.ExpiresAt) {
			_ = uc.refreshRepo.Revoke(ctx, stored.ID)
		}
		return Output{}, domain.ErrRefreshTokenInvalid
	}

	accessToken, err := uc.access.Issue(stored.UserID.String(), uc.accessTTL)
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}

	newRefresh, err := common.NewRefreshToken()
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}
	newHash := common.HashToken(newRefresh)
	if err := uc.uow.Do(ctx, func(ctx context.Context) error {
		if err := uc.refreshRepo.Revoke(ctx, stored.ID); err != nil {
			return common.NormalizeError(err)
		}
		return common.NormalizeError(
			uc.refreshRepo.Create(ctx, domain.NewRefreshTokenRecord(stored.UserID, newHash, now, uc.refreshTTL)),
		)
	}); err != nil {
		return Output{}, err
	}

	return Output{
		AccessToken:  accessToken,
		RefreshToken: newRefresh,
	}, nil
}
