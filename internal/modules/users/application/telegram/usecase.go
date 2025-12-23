package telegram

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
)

const telegramProvider = "telegram"

type Input struct {
	InitData string `json:"init_data"`
}

type UseCase struct {
	users      domain.UserRepository
	identities domain.IdentityRepository
	refresh    domain.RefreshTokenRepository

	access     common.AccessTokenIssuer
	accessTTL  time.Duration
	refreshTTL time.Duration

	validator validator
}

type validator struct {
	botToken    string
	initDataTTL time.Duration
}

type telegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
}

func New(
	users domain.UserRepository,
	identities domain.IdentityRepository,
	refresh domain.RefreshTokenRepository,
	access common.AccessTokenIssuer,
	botToken string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	initDataTTL time.Duration,
) (*UseCase, error) {
	if botToken = strings.TrimSpace(botToken); botToken == "" {
		return nil, domain.ErrUnauthorized
	}

	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL == 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	if initDataTTL == 0 {
		initDataTTL = 24 * time.Hour
	}

	return &UseCase{
		users:      users,
		identities: identities,
		refresh:    refresh,
		access:     access,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		validator: validator{
			botToken:    botToken,
			initDataTTL: initDataTTL,
		},
	}, nil
}

func (uc *UseCase) Execute(ctx context.Context, in Input) (login.Output, error) {
	payload, err := uc.validator.parse(in.InitData)
	if err != nil {
		return login.Output{}, domain.ErrInvalidCredentials
	}

	providerUserID := strconv.FormatInt(payload.ID, 10)

	ident, found, err := uc.identities.GetByProvider(ctx, telegramProvider, providerUserID)
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
		user, err = uc.registerUser(ctx, payload)
		if err != nil {
			return login.Output{}, common.NormalizeError(err)
		}

		identity, err := domain.NewExternalIdentity(user.ID, telegramProvider, providerUserID, time.Now().UTC())
		if err != nil {
			return login.Output{}, err
		}
		if err := domain.EnsureIdentityAvailable(ctx, uc.identities, user.ID, telegramProvider, providerUserID); err != nil {
			return login.Output{}, common.NormalizeError(err)
		}
		if err := uc.identities.Create(ctx, identity); err != nil {
			return login.Output{}, common.NormalizeError(err)
		}
	}

	return uc.issueTokens(ctx, user)
}

func (uc *UseCase) registerUser(ctx context.Context, payload telegramUser) (domain.User, error) {
	displayName, err := uc.displayName(payload)
	if err != nil {
		return domain.User{}, err
	}

	userID := domain.NewUserID()
	now := time.Now().UTC()
	user := domain.NewUser(userID, "", displayName, now)
	user.FirstName = strings.TrimSpace(payload.FirstName)
	user.LastName = strings.TrimSpace(payload.LastName)

	if payload.PhotoURL != "" {
		if avatar, err := domain.NewAvatarURL(payload.PhotoURL); err == nil {
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
	refreshRecord, reuse, err := common.PrepareRefreshRecord(ctx, uc.refresh, user.ID, refreshHash, now, uc.refreshTTL)
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}

	accessToken, err := uc.access.Issue(user.ID.String(), refreshRecord.ID, uc.accessTTL)
	if err != nil {
		return login.Output{}, common.NormalizeError(err)
	}
	if reuse {
		if err := uc.refresh.Update(ctx, refreshRecord); err != nil {
			return login.Output{}, common.NormalizeError(err)
		}
	} else if err := uc.refresh.Create(ctx, refreshRecord); err != nil {
		return login.Output{}, common.NormalizeError(err)
	}

	return login.Output{
		UserID:       user.ID.String(),
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		MiddleName:   user.MiddleName,
		DisplayName:  user.DisplayName,
		AvatarURL:    user.AvatarURL,
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
	}, nil
}

func (uc *UseCase) displayName(payload telegramUser) (domain.DisplayName, error) {
	candidates := []string{
		payload.Username,
		strings.TrimSpace(strings.Join([]string{payload.FirstName, payload.LastName}, " ")),
	}
	for _, candidate := range candidates {
		if displayName, err := domain.NewDisplayName(candidate); err == nil {
			return displayName, nil
		}
	}
	fallback := "tg_" + strconv.FormatInt(payload.ID, 10)
	return domain.NewDisplayName(fallback)
}

func (v validator) parse(raw string) (telegramUser, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return telegramUser{}, domain.ErrInvalidCredentials
	}

	values, err := url.ParseQuery(raw)
	if err != nil {
		return telegramUser{}, err
	}

	hash := values.Get("hash")
	if hash == "" {
		return telegramUser{}, domain.ErrInvalidCredentials
	}
	values.Del("hash")

	if !v.verifyHash(values, hash) {
		return telegramUser{}, domain.ErrInvalidCredentials
	}

	authDateRaw := values.Get("auth_date")
	authDate, err := strconv.ParseInt(authDateRaw, 10, 64)
	if err != nil {
		return telegramUser{}, domain.ErrInvalidCredentials
	}
	if v.initDataTTL > 0 && time.Since(time.Unix(authDate, 0)) > v.initDataTTL {
		return telegramUser{}, domain.ErrInvalidCredentials
	}

	var user telegramUser
	if err := json.Unmarshal([]byte(values.Get("user")), &user); err != nil {
		return telegramUser{}, domain.ErrInvalidCredentials
	}
	if user.ID == 0 {
		return telegramUser{}, domain.ErrInvalidCredentials
	}
	return user, nil
}

func (v validator) verifyHash(values url.Values, expected string) bool {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var pairs []string
	for _, key := range keys {
		pairs = append(pairs, key+"="+values.Get(key))
	}
	dataCheckString := strings.Join(pairs, "\n")

	secretHasher := hmac.New(sha256.New, []byte("WebAppData"))
	secretHasher.Write([]byte(v.botToken))
	secret := secretHasher.Sum(nil)

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(dataCheckString))
	calculated := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(calculated), []byte(expected))
}
