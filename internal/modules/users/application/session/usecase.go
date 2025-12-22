package session

import (
	"context"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type UseCase struct {
	refresh domain.RefreshTokenRepository
}

func New(refresh domain.RefreshTokenRepository) *UseCase {
	return &UseCase{refresh: refresh}
}

func (uc *UseCase) List(ctx context.Context, in ListInput) (Output, error) {
	userID, err := domain.ParseUserID(in.UserID)
	if err != nil {
		return Output{}, err
	}

	currentID := ""
	if in.CurrentRefreshToken != "" {
		hash := common.HashToken(in.CurrentRefreshToken)
		current, found, err := uc.refresh.GetByHash(ctx, hash)
		if err != nil {
			return Output{}, common.NormalizeError(err)
		}
		if !found || current.UserID != userID || !current.IsValid(time.Now().UTC()) {
			return Output{}, domain.ErrRefreshTokenInvalid
		}
		currentID = current.ID
	}

	tokens, err := uc.refresh.ListByUser(ctx, userID)
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}

	sessions := make([]Session, 0, len(tokens))
	for _, t := range tokens {
		session := Session{
			ID:        t.ID,
			UserAgent: t.UserAgent,
			IP:        t.IP,
			CreatedAt: t.CreatedAt,
			ExpiresAt: t.ExpiresAt,
			Current:   t.ID == currentID,
		}
		if t.RevokedAt != nil {
			session.RevokedAt = t.RevokedAt
		}
		sessions = append(sessions, session)
	}

	return Output{Sessions: sessions}, nil
}

func (uc *UseCase) Revoke(ctx context.Context, in RevokeInput) (struct{}, error) {
	userID, err := domain.ParseUserID(in.UserID)
	if err != nil {
		return struct{}{}, err
	}

	token, found, err := uc.refresh.GetByID(ctx, in.SessionID)
	if err != nil {
		return struct{}{}, common.NormalizeError(err)
	}
	if !found || token.UserID != userID {
		return struct{}{}, domain.ErrUnauthorized
	}

	if err := uc.refresh.Revoke(ctx, in.SessionID); err != nil {
		return struct{}{}, common.NormalizeError(err)
	}

	return struct{}{}, nil
}

func (uc *UseCase) RevokeOthers(ctx context.Context, in RevokeOthersInput) (struct{}, error) {
	if in.CurrentRefreshToken == "" {
		return struct{}{}, domain.ErrRefreshTokenInvalid
	}
	userID, err := domain.ParseUserID(in.UserID)
	if err != nil {
		return struct{}{}, err
	}

	hash := common.HashToken(in.CurrentRefreshToken)
	current, found, err := uc.refresh.GetByHash(ctx, hash)
	if err != nil {
		return struct{}{}, common.NormalizeError(err)
	}
	if !found || current.UserID != userID || !current.IsValid(time.Now().UTC()) {
		return struct{}{}, domain.ErrRefreshTokenInvalid
	}

	if err := uc.refresh.RevokeAllExcept(ctx, userID, []string{current.ID}); err != nil {
		return struct{}{}, common.NormalizeError(err)
	}

	return struct{}{}, nil
}
