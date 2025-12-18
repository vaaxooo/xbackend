package link

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

type UseCase struct {
	identities domain.IdentityRepository
}

func New(identities domain.IdentityRepository) *UseCase {
	return &UseCase{identities: identities}
}

func (uc *UseCase) Execute(ctx context.Context, in Input) (Output, error) {
	if in.UserID == "" {
		return Output{}, domain.ErrUnauthorized
	}
	if in.Provider == "" || in.ProviderUserID == "" {
		return Output{}, domain.ErrInvalidCredentials
	}

	// Prevent linking the same provider twice for the same user.
	if _, found, err := uc.identities.GetByUserAndProvider(ctx, in.UserID, in.Provider); err != nil {
		return Output{}, err
	} else if found {
		return Output{}, domain.ErrIdentityAlreadyLinked
	}

	// Prevent linking an identity that already belongs to another user.
	if _, found, err := uc.identities.GetByProvider(ctx, in.Provider, in.ProviderUserID); err != nil {
		return Output{}, err
	} else if found {
		return Output{}, domain.ErrIdentityAlreadyLinked
	}

	if err := uc.identities.Create(ctx, domain.Identity{
		ID:             uuid.NewString(),
		UserID:         in.UserID,
		Provider:       in.Provider,
		ProviderUserID: in.ProviderUserID,
		SecretHash:     "",
		CreatedAt:      time.Now().UTC(),
	}); err != nil {
		return Output{}, err
	}

	return Output{Linked: true}, nil
}
