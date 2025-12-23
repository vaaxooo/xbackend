package google

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

const providerName = "google"

type Input struct {
	IDToken string `json:"id_token"`
}

type claims struct {
	Email         string `json:"email"`
	EmailVerified any    `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	jwt.RegisteredClaims
}

type tokenVerifier interface {
	Verify(ctx context.Context, raw string, claims jwt.Claims) error
}

type UseCase struct {
	users      domain.UserRepository
	identities domain.IdentityRepository
	refresh    domain.RefreshTokenRepository
	access     common.AccessTokenIssuer
	accessTTL  time.Duration
	refreshTTL time.Duration
	verifier   tokenVerifier
}

func New(
	users domain.UserRepository,
	identities domain.IdentityRepository,
	refresh domain.RefreshTokenRepository,
	access common.AccessTokenIssuer,
	verifier tokenVerifier,
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
		users:      users,
		identities: identities,
		refresh:    refresh,
		access:     access,
		verifier:   verifier,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (uc *UseCase) Execute(ctx context.Context, in Input) (login.Output, error) {
	var tokenClaims claims
	if err := uc.verifier.Verify(ctx, in.IDToken, &tokenClaims); err != nil {
		return login.Output{}, domain.ErrInvalidCredentials
	}
	if tokenClaims.Subject == "" || tokenClaims.Email == "" {
		return login.Output{}, domain.ErrInvalidCredentials
	}

	ident, found, err := uc.identities.GetByProvider(ctx, providerName, tokenClaims.Subject)
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}

	var user domain.User
	if found {
		user, found, err = uc.users.GetByID(ctx, ident.UserID)
		if err != nil {
			return login.Output{}, common.NormalizeError(err)
		}
		if !found {
			return login.Output{}, domain.ErrInvalidCredentials
		}
	} else {
		user, err = uc.registerUser(ctx, tokenClaims)
		if err != nil {
			return login.Output{}, common.NormalizeError(err)
		}

		identity, err := domain.NewExternalIdentity(user.ID, providerName, tokenClaims.Subject, time.Now().UTC())
		if err != nil {
			return login.Output{}, err
		}
		if verified(tokenClaims.EmailVerified) {
			identity = identity.WithEmailVerified(time.Now().UTC())
		}
		if err := domain.EnsureIdentityAvailable(ctx, uc.identities, user.ID, providerName, tokenClaims.Subject); err != nil {
			return login.Output{}, common.NormalizeError(err)
		}
		if err := uc.identities.Create(ctx, identity); err != nil {
			return login.Output{}, common.NormalizeError(err)
		}
	}

	return uc.issueTokens(ctx, user)
}

func (uc *UseCase) registerUser(ctx context.Context, c claims) (domain.User, error) {
	email, err := domain.NewEmail(c.Email)
	if err != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}

	displayName, err := uc.displayName(c)
	if err != nil {
		return domain.User{}, err
	}

	userID := domain.NewUserID()
	now := time.Now().UTC()
	user := domain.NewUser(userID, email.String(), displayName, now)
	user.FirstName = strings.TrimSpace(c.GivenName)
	user.LastName = strings.TrimSpace(c.FamilyName)
	user.DisplayName = displayName.String()

	if c.Picture != "" {
		if avatar, err := domain.NewAvatarURL(c.Picture); err == nil {
			user.AvatarURL = avatar.String()
		}
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (uc *UseCase) issueTokens(ctx context.Context, user domain.User) (login.Output, error) {
	refreshRaw, err := common.NewRefreshToken()
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}
	refreshHash := common.HashToken(refreshRaw)
	now := time.Now().UTC()
	refreshRecord := common.NewRefreshRecord(ctx, user.ID, refreshHash, now, uc.refreshTTL)

	accessToken, err := uc.access.Issue(user.ID.String(), refreshRecord.ID, uc.accessTTL)
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}
	if err := uc.refresh.Create(ctx, refreshRecord); err != nil {
		return login.Output{}, common.NormalizeError(err)
	}

	return login.Output{
		UserID:       user.ID.String(),
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		MiddleName:   user.MiddleName,
		DisplayName:  user.DisplayName,
		AvatarURL:    user.AvatarURL,
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
	}, nil
}

func (uc *UseCase) displayName(c claims) (domain.DisplayName, error) {
	candidates := []string{c.Name, strings.TrimSpace(strings.Join([]string{c.GivenName, c.FamilyName}, " "))}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if displayName, err := domain.NewDisplayName(candidate); err == nil {
			return displayName, nil
		}
	}
	return domain.NewDisplayName(providerName + "_" + c.Subject)
}

func verified(val any) bool {
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true")
	default:
		return false
	}
}
