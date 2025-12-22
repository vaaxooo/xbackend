package profile

import (
	"context"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type GetUseCase struct {
	users domain.UserRepository
}

type UpdateUseCase struct {
	users domain.UserRepository
}

func NewGet(users domain.UserRepository) *GetUseCase {
	return &GetUseCase{users: users}
}

func NewUpdate(users domain.UserRepository) *UpdateUseCase {
	return &UpdateUseCase{users: users}
}

func (uc *GetUseCase) Execute(ctx context.Context, in GetInput) (Output, error) {
	id, err := domain.ParseUserID(in.UserID)
	if err != nil {
		return Output{}, err
	}

	u, ok, err := uc.users.GetByID(ctx, id)
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}
	if !ok {
		return Output{}, domain.ErrUnauthorized
	}

	return Output{
		UserID:      u.ID.String(),
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		MiddleName:  u.MiddleName,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
	}, nil
}

func (uc *UpdateUseCase) Execute(ctx context.Context, in UpdateInput) (Output, error) {
	id, err := domain.ParseUserID(in.UserID)
	if err != nil {
		return Output{}, err
	}

	// PATCH semantics:
	// - if a field is not provided (nil), we keep the current value
	// - if provided as an empty string, we treat it as a request to clear the value
	current, ok, err := uc.users.GetByID(ctx, id)
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}
	if !ok {
		return Output{}, domain.ErrUnauthorized
	}

	patched, err := current.ApplyPatch(domain.ProfilePatch{
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		MiddleName:  in.MiddleName,
		DisplayName: in.DisplayName,
		AvatarURL:   in.AvatarURL,
	})
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}

	u, err := uc.users.UpdateProfile(ctx, patched)
	if err != nil {
		return Output{}, common.NormalizeError(err)
	}

	return Output{
		UserID:      u.ID.String(),
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		MiddleName:  u.MiddleName,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
	}, nil
}
