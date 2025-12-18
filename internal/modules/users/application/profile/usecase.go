package profile

import (
	"context"
	"strings"

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
	if in.UserID == "" {
		return Output{}, domain.ErrUnauthorized
	}

	u, ok, err := uc.users.GetByID(ctx, in.UserID)
	if err != nil {
		return Output{}, err
	}
	if !ok {
		return Output{}, domain.ErrUnauthorized
	}

	return Output{
		UserID:      u.ID,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		MiddleName:  u.MiddleName,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
	}, nil
}

func (uc *UpdateUseCase) Execute(ctx context.Context, in UpdateInput) (Output, error) {
	if in.UserID == "" {
		return Output{}, domain.ErrUnauthorized
	}

	// PATCH semantics:
	// - if a field is not provided (nil), we keep the current value
	// - if provided as an empty string, we treat it as a request to clear the value
	current, ok, err := uc.users.GetByID(ctx, in.UserID)
	if err != nil {
		return Output{}, err
	}
	if !ok {
		return Output{}, domain.ErrUnauthorized
	}

	if in.FirstName != nil {
		current.FirstName = strings.TrimSpace(*in.FirstName)
	}
	if in.LastName != nil {
		current.LastName = strings.TrimSpace(*in.LastName)
	}
	if in.MiddleName != nil {
		current.MiddleName = strings.TrimSpace(*in.MiddleName)
	}
	if in.DisplayName != nil {
		current.DisplayName = strings.TrimSpace(*in.DisplayName)
	}
	if in.AvatarURL != nil {
		current.AvatarURL = strings.TrimSpace(*in.AvatarURL)
	}

	u, err := uc.users.UpdateProfile(ctx, current)
	if err != nil {
		return Output{}, err
	}

	return Output{
		UserID:      u.ID,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		MiddleName:  u.MiddleName,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
	}, nil
}
