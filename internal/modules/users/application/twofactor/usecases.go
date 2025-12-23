package twofactor

import (
	"context"
	"strings"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/totp"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type SetupInput struct {
	UserID string
}

type SetupOutput struct {
	Secret         string
	ProvisioningQR string
}

type ConfirmInput struct {
	UserID string
	Code   string
}

type DisableInput struct {
	UserID string
	Code   string
}

type UseCase struct {
	identities domain.IdentityRepository
	issuer     string
}

func NewUseCase(identities domain.IdentityRepository, issuer string) *UseCase {
	if issuer == "" {
		issuer = "xbackend"
	}
	return &UseCase{identities: identities, issuer: issuer}
}

func (uc *UseCase) Setup(ctx context.Context, in SetupInput) (SetupOutput, error) {
	userID, err := domain.ParseUserID(in.UserID)
	if err != nil {
		return SetupOutput{}, err
	}

	ident, found, err := uc.identities.GetByUserAndProvider(ctx, userID, "email")
	if err != nil {
		return SetupOutput{}, common.NormalizeError(err)
	}
	if !found {
		return SetupOutput{}, domain.ErrInvalidCredentials
	}
	if ident.IsTwoFactorEnabled() {
		return SetupOutput{}, domain.ErrTwoFactorAlreadyEnabled
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      uc.issuer,
		AccountName: ident.ProviderUserID,
	})
	if err != nil {
		return SetupOutput{}, common.NormalizeError(err)
	}

	ident = ident.WithTOTPSecret(key.Secret())
	if err := uc.identities.Update(ctx, ident); err != nil {
		return SetupOutput{}, common.NormalizeError(err)
	}

	return SetupOutput{Secret: key.Secret(), ProvisioningQR: key.URL()}, nil
}

func (uc *UseCase) Confirm(ctx context.Context, in ConfirmInput) error {
	userID, err := domain.ParseUserID(in.UserID)
	if err != nil {
		return err
	}

	ident, found, err := uc.identities.GetByUserAndProvider(ctx, userID, "email")
	if err != nil {
		return common.NormalizeError(err)
	}
	if !found || strings.TrimSpace(ident.TOTPSecret) == "" {
		return domain.ErrInvalidCredentials
	}

	if !totp.Validate(in.Code, ident.TOTPSecret) {
		return domain.ErrInvalidTwoFactor
	}

	ident = ident.WithTOTPConfirmed(time.Now().UTC())
	return uc.identities.Update(ctx, ident)
}

func (uc *UseCase) Disable(ctx context.Context, in DisableInput) error {
	userID, err := domain.ParseUserID(in.UserID)
	if err != nil {
		return err
	}

	ident, found, err := uc.identities.GetByUserAndProvider(ctx, userID, "email")
	if err != nil {
		return common.NormalizeError(err)
	}
	if !found || !ident.IsTwoFactorEnabled() {
		return domain.ErrInvalidCredentials
	}

	if !totp.Validate(in.Code, ident.TOTPSecret) {
		return domain.ErrInvalidTwoFactor
	}

	ident = ident.ClearTOTP()
	return uc.identities.Update(ctx, ident)
}
