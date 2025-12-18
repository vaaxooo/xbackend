package link

import (
	"context"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type UseCase struct {
	identities domain.IdentityRepository
}

func New(identities domain.IdentityRepository) *UseCase {
	return &UseCase{identities: identities}
}

func (uc *UseCase) Execute(ctx context.Context, in Input) (Output, error) {
	userID, err := domain.ParseUserID(in.UserID)
	if err != nil {
		return Output{}, err
	}

	if err := domain.EnsureIdentityAvailable(ctx, uc.identities, userID, in.Provider, in.ProviderUserID); err != nil {
		return Output{}, err
	}

	identity, err := domain.NewExternalIdentity(userID, in.Provider, in.ProviderUserID, time.Now().UTC())
	if err != nil {
		return Output{}, err
	}

	if err := uc.identities.Create(ctx, identity); err != nil {
		return Output{}, err
	}

	return Output{Linked: true}, nil
}
