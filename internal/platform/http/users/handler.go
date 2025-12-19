package users

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/profile"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/refresh"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
	usersapi "github.com/vaaxooo/xbackend/internal/modules/users/public"
	"github.com/vaaxooo/xbackend/internal/platform/http/users/dto"
	"github.com/vaaxooo/xbackend/internal/platform/http/users/httpctx"

	phttp "github.com/vaaxooo/xbackend/internal/platform/http"
)

type Handler struct {
	middleware phttp.UseCaseMiddleware

	register phttp.UseCaseHandler[usersapi.RegisterInput, login.Output]
	login    phttp.UseCaseHandler[usersapi.LoginInput, login.Output]
	telegram phttp.UseCaseHandler[usersapi.TelegramLoginInput, login.Output]
	refresh  phttp.UseCaseHandler[usersapi.RefreshInput, refresh.Output]

	getMe  phttp.UseCaseHandler[usersapi.GetProfileInput, profile.Output]
	update phttp.UseCaseHandler[usersapi.UpdateProfileInput, profile.Output]
	link   phttp.UseCaseHandler[usersapi.LinkProviderInput, link.Output]
}

func NewHandler(svc usersapi.Service, middleware phttp.UseCaseMiddleware) *Handler {
	return &Handler{
		middleware: middleware,
		register: phttp.UseCaseFunc[usersapi.RegisterInput, login.Output](func(ctx context.Context, cmd usersapi.RegisterInput) (login.Output, error) {
			return svc.Register(ctx, cmd)
		}),
		login: phttp.UseCaseFunc[usersapi.LoginInput, login.Output](func(ctx context.Context, cmd usersapi.LoginInput) (login.Output, error) {
			return svc.Login(ctx, cmd)
		}),
		telegram: phttp.UseCaseFunc[usersapi.TelegramLoginInput, login.Output](func(ctx context.Context, cmd usersapi.TelegramLoginInput) (login.Output, error) {
			return svc.LoginWithTelegram(ctx, cmd)
		}),
		refresh: phttp.UseCaseFunc[usersapi.RefreshInput, refresh.Output](func(ctx context.Context, cmd usersapi.RefreshInput) (refresh.Output, error) {
			return svc.Refresh(ctx, cmd)
		}),
		getMe: phttp.UseCaseFunc[usersapi.GetProfileInput, profile.Output](func(ctx context.Context, cmd usersapi.GetProfileInput) (profile.Output, error) {
			return svc.GetMe(ctx, cmd)
		}),
		update: phttp.UseCaseFunc[usersapi.UpdateProfileInput, profile.Output](func(ctx context.Context, cmd usersapi.UpdateProfileInput) (profile.Output, error) {
			return svc.UpdateProfile(ctx, cmd)
		}),
		link: phttp.UseCaseFunc[usersapi.LinkProviderInput, link.Output](func(ctx context.Context, cmd usersapi.LinkProviderInput) (link.Output, error) {
			return svc.LinkProvider(ctx, cmd)
		}),
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}

	out, err := phttp.HandleUseCase(h.middleware, r, h.register, usersapi.RegisterInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}

	phttp.WriteJSON(w, http.StatusCreated, dto.LoginResponse{
		UserProfileResponse: dto.UserProfileResponse{
			UserID:      out.UserID,
			FirstName:   out.FirstName,
			LastName:    out.LastName,
			MiddleName:  out.MiddleName,
			DisplayName: out.DisplayName,
			AvatarURL:   out.AvatarURL,
		},
		TokensResponse: dto.TokensResponse{
			AccessToken:  out.AccessToken,
			RefreshToken: out.RefreshToken,
		},
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}

	out, err := phttp.HandleUseCase(h.middleware, r, h.login, usersapi.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}

	phttp.WriteJSON(w, http.StatusOK, dto.LoginResponse{
		UserProfileResponse: dto.UserProfileResponse{
			UserID:      out.UserID,
			FirstName:   out.FirstName,
			LastName:    out.LastName,
			MiddleName:  out.MiddleName,
			DisplayName: out.DisplayName,
			AvatarURL:   out.AvatarURL,
		},
		TokensResponse: dto.TokensResponse{
			AccessToken:  out.AccessToken,
			RefreshToken: out.RefreshToken,
		},
	})
}

func (h *Handler) TelegramLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.TelegramLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}

	out, err := phttp.HandleUseCase(h.middleware, r, h.telegram, usersapi.TelegramLoginInput{
		InitData: req.InitData,
	})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}

	phttp.WriteJSON(w, http.StatusOK, dto.LoginResponse{
		UserProfileResponse: dto.UserProfileResponse{
			UserID:      out.UserID,
			FirstName:   out.FirstName,
			LastName:    out.LastName,
			MiddleName:  out.MiddleName,
			DisplayName: out.DisplayName,
			AvatarURL:   out.AvatarURL,
		},
		TokensResponse: dto.TokensResponse{
			AccessToken:  out.AccessToken,
			RefreshToken: out.RefreshToken,
		},
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}

	out, err := phttp.HandleUseCase(h.middleware, r, h.refresh, usersapi.RefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}

	phttp.WriteJSON(w, http.StatusOK, dto.TokensResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	uid, ok := httpctx.UserIDFromContext(r.Context())
	if !ok || uid == "" {
		phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	out, err := phttp.HandleUseCase(h.middleware, r, h.getMe, usersapi.GetProfileInput{
		UserID: uid,
	})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}

	phttp.WriteJSON(w, http.StatusOK, dto.UserProfileResponse{
		UserID:      out.UserID,
		FirstName:   out.FirstName,
		LastName:    out.LastName,
		MiddleName:  out.MiddleName,
		DisplayName: out.DisplayName,
		AvatarURL:   out.AvatarURL,
	})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	uid, ok := httpctx.UserIDFromContext(r.Context())
	if !ok {
		phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}

	out, err := phttp.HandleUseCase(h.middleware, r, h.update, usersapi.UpdateProfileInput{
		UserID:      uid,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		MiddleName:  req.MiddleName,
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarURL,
	})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}

	phttp.WriteJSON(w, http.StatusOK, dto.UserProfileResponse{
		UserID:      out.UserID,
		FirstName:   out.FirstName,
		LastName:    out.LastName,
		MiddleName:  out.MiddleName,
		DisplayName: out.DisplayName,
		AvatarURL:   out.AvatarURL,
	})
}

func (h *Handler) LinkProvider(w http.ResponseWriter, r *http.Request) {
	uid, ok := httpctx.UserIDFromContext(r.Context())
	if !ok {
		phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	var req dto.LinkProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}

	out, err := phttp.HandleUseCase(h.middleware, r, h.link, usersapi.LinkProviderInput{
		UserID:         uid,
		Provider:       req.Provider,
		ProviderUserID: req.ProviderUserID,
	})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}

	phttp.WriteJSON(w, http.StatusOK, dto.LinkProviderResponse{
		Linked: out.Linked,
	})
}

func mapError(err error) (status int, code string, message string) {
	if errors.Is(err, domain.ErrInvalidEmail) || errors.Is(err, domain.ErrWeakPassword) {
		return http.StatusBadRequest, "validation_error", "Validation error"
	}
	if errors.Is(err, domain.ErrEmailAlreadyUsed) || errors.Is(err, domain.ErrIdentityAlreadyLinked) {
		return http.StatusConflict, "conflict", "Conflict"
	}
	if errors.Is(err, domain.ErrInvalidCredentials) || errors.Is(err, domain.ErrRefreshTokenInvalid) {
		return http.StatusUnauthorized, "unauthorized", "Unauthorized"
	}
	if errors.Is(err, domain.ErrUnauthorized) {
		return http.StatusUnauthorized, "unauthorized", "Unauthorized"
	}
	if errors.Is(err, common.ErrInternal) {
		return http.StatusInternalServerError, "internal_error", "Internal server error"
	}
	return http.StatusInternalServerError, "internal_error", "Internal server error"
}
