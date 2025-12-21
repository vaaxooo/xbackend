package password

import (
	"context"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type ChangeInput struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}

type ChangeUseCase struct {
	identities domain.IdentityRepository
	hasher     domain.PasswordHasher
}

func NewChange(identities domain.IdentityRepository, hasher domain.PasswordHasher) *ChangeUseCase {
	return &ChangeUseCase{identities: identities, hasher: hasher}
}

func (uc *ChangeUseCase) Execute(ctx context.Context, in ChangeInput) (struct{}, error) {
	userID, err := domain.ParseUserID(in.UserID)
	if err != nil {
		return struct{}{}, err
	}

	ident, found, err := uc.identities.GetByUserAndProvider(ctx, userID, "email")
	if err != nil {
		return struct{}{}, common.NormalizeError(err)
	}
	if !found || ident.SecretHash == "" {
		return struct{}{}, domain.ErrInvalidCredentials
	}

	if err := ident.Authenticate(ctx, uc.hasher, in.CurrentPassword); err != nil {
		return struct{}{}, domain.ErrInvalidCredentials
	}

	newHash, err := domain.NewPasswordHash(ctx, in.NewPassword, uc.hasher)
	if err != nil {
		return struct{}{}, common.NormalizeError(err)
	}

	ident.SecretHash = newHash
	if err := uc.identities.Update(ctx, ident); err != nil {
		return struct{}{}, common.NormalizeError(err)
	}

	return struct{}{}, nil
}
