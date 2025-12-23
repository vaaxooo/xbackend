package profile

import (
	"context"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type GetUseCase struct {
	users      domain.UserRepository
	identities domain.IdentityRepository
}

type UpdateUseCase struct {
	users      domain.UserRepository
	identities domain.IdentityRepository
}

func NewGet(users domain.UserRepository, identities domain.IdentityRepository) *GetUseCase {
	return &GetUseCase{users: users, identities: identities}
}

func NewUpdate(users domain.UserRepository, identities domain.IdentityRepository) *UpdateUseCase {
	return &UpdateUseCase{users: users, identities: identities}
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

	settings, err := loginSettings(ctx, uc.identities, u.ID)
	if err != nil {
		return Output{}, err
	}

	return Output{
		UserID:        u.ID.String(),
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		MiddleName:    u.MiddleName,
		DisplayName:   u.DisplayName,
		AvatarURL:     u.AvatarURL,
		LoginSettings: settings,
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

	settings, err := loginSettings(ctx, uc.identities, u.ID)
	if err != nil {
		return Output{}, err
	}

	return Output{
		UserID:        u.ID.String(),
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		MiddleName:    u.MiddleName,
		DisplayName:   u.DisplayName,
		AvatarURL:     u.AvatarURL,
		LoginSettings: settings,
	}, nil
}

func loginSettings(ctx context.Context, repo domain.IdentityRepository, userID domain.UserID) (LoginSettings, error) {
	ident, found, err := repo.GetByUserAndProvider(ctx, userID, "email")
	if err != nil {
		return LoginSettings{}, common.NormalizeError(err)
	}
	if !found {
		return LoginSettings{}, domain.ErrUnauthorized
	}

	return LoginSettings{
		TwoFactorEnabled: ident.IsTwoFactorEnabled(),
		EmailVerified:    ident.IsEmailVerified(),
	}, nil
}
