package users

import (
	"encoding/json"
	"errors"
	"net/http"

	usersapp "github.com/vaaxooo/xbackend/internal/modules/users/application"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/profile"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/refresh"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/register"
	"github.com/vaaxooo/xbackend/internal/modules/users/domain"
	"github.com/vaaxooo/xbackend/internal/platform/http/users/dto"
	"github.com/vaaxooo/xbackend/internal/platform/http/users/httpctx"

	phttp "github.com/vaaxooo/xbackend/internal/platform/http"
)

type Handler struct {
	svc usersapp.Service
}

func NewHandler(svc usersapp.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}

	out, err := h.svc.Register(r.Context(), register.Input{
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

	out, err := h.svc.Login(r.Context(), login.Input{
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

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}

	out, err := h.svc.Refresh(r.Context(), refresh.Input{
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

	out, err := h.svc.GetMe(r.Context(), profile.GetInput{
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

	out, err := h.svc.UpdateProfile(r.Context(), profile.UpdateInput{
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

	out, err := h.svc.LinkProvider(r.Context(), link.Input{
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
