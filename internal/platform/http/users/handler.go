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

	register              phttp.UseCaseHandler[usersapi.RegisterInput, login.Output]
	login                 phttp.UseCaseHandler[usersapi.LoginInput, login.Output]
	telegram              phttp.UseCaseHandler[usersapi.TelegramLoginInput, login.Output]
	google                phttp.UseCaseHandler[usersapi.GoogleLoginInput, login.Output]
	apple                 phttp.UseCaseHandler[usersapi.AppleLoginInput, login.Output]
	refresh               phttp.UseCaseHandler[usersapi.RefreshInput, refresh.Output]
	confirmEmail          phttp.UseCaseHandler[usersapi.ConfirmEmailInput, login.Output]
	requestConfirm        phttp.UseCaseHandler[usersapi.RequestEmailInput, struct{}]
	requestPasswordReset  phttp.UseCaseHandler[usersapi.RequestPasswordResetInput, struct{}]
	resetPassword         phttp.UseCaseHandler[usersapi.ResetPasswordInput, struct{}]
	setupTwoFactor        phttp.UseCaseHandler[usersapi.TwoFactorSetupInput, usersapi.TwoFactorSetupOutput]
	confirmTwoFactor      phttp.UseCaseHandler[usersapi.TwoFactorConfirmInput, struct{}]
	disableTwoFactor      phttp.UseCaseHandler[usersapi.TwoFactorDisableInput, struct{}]
	challengeStatus       phttp.UseCaseHandler[usersapi.ChallengeStatusInput, login.Output]
	challengeVerifyTOTP   phttp.UseCaseHandler[usersapi.ChallengeVerifyTOTPInput, login.Output]
	challengeResendEmail  phttp.UseCaseHandler[usersapi.ChallengeResendEmailInput, login.Output]
	challengeConfirmEmail phttp.UseCaseHandler[usersapi.ChallengeConfirmEmailInput, login.Output]

	getMe          phttp.UseCaseHandler[usersapi.GetProfileInput, profile.Output]
	update         phttp.UseCaseHandler[usersapi.UpdateProfileInput, profile.Output]
	changePassword phttp.UseCaseHandler[usersapi.ChangePasswordInput, struct{}]
	link           phttp.UseCaseHandler[usersapi.LinkProviderInput, link.Output]

	listSessions       phttp.UseCaseHandler[usersapi.ListSessionsInput, usersapi.SessionsOutput]
	revokeSession      phttp.UseCaseHandler[usersapi.RevokeSessionInput, struct{}]
	revokeOtherSession phttp.UseCaseHandler[usersapi.RevokeOtherSessionsInput, struct{}]
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
		google: phttp.UseCaseFunc[usersapi.GoogleLoginInput, login.Output](func(ctx context.Context, cmd usersapi.GoogleLoginInput) (login.Output, error) {
			return svc.LoginWithGoogle(ctx, cmd)
		}),
		apple: phttp.UseCaseFunc[usersapi.AppleLoginInput, login.Output](func(ctx context.Context, cmd usersapi.AppleLoginInput) (login.Output, error) {
			return svc.LoginWithApple(ctx, cmd)
		}),
		refresh: phttp.UseCaseFunc[usersapi.RefreshInput, refresh.Output](func(ctx context.Context, cmd usersapi.RefreshInput) (refresh.Output, error) {
			return svc.Refresh(ctx, cmd)
		}),
		confirmEmail: phttp.UseCaseFunc[usersapi.ConfirmEmailInput, login.Output](func(ctx context.Context, cmd usersapi.ConfirmEmailInput) (login.Output, error) {
			return svc.ConfirmEmail(ctx, cmd)
		}),
		requestConfirm: phttp.UseCaseFunc[usersapi.RequestEmailInput, struct{}](func(ctx context.Context, cmd usersapi.RequestEmailInput) (struct{}, error) {
			return struct{}{}, svc.RequestEmailConfirmation(ctx, cmd)
		}),
		requestPasswordReset: phttp.UseCaseFunc[usersapi.RequestPasswordResetInput, struct{}](func(ctx context.Context, cmd usersapi.RequestPasswordResetInput) (struct{}, error) {
			return struct{}{}, svc.RequestPasswordReset(ctx, cmd)
		}),
		resetPassword: phttp.UseCaseFunc[usersapi.ResetPasswordInput, struct{}](func(ctx context.Context, cmd usersapi.ResetPasswordInput) (struct{}, error) {
			return struct{}{}, svc.ResetPassword(ctx, cmd)
		}),
		setupTwoFactor: phttp.UseCaseFunc[usersapi.TwoFactorSetupInput, usersapi.TwoFactorSetupOutput](func(ctx context.Context, cmd usersapi.TwoFactorSetupInput) (usersapi.TwoFactorSetupOutput, error) {
			return svc.SetupTwoFactor(ctx, cmd)
		}),
		confirmTwoFactor: phttp.UseCaseFunc[usersapi.TwoFactorConfirmInput, struct{}](func(ctx context.Context, cmd usersapi.TwoFactorConfirmInput) (struct{}, error) {
			return struct{}{}, svc.ConfirmTwoFactor(ctx, cmd)
		}),
		disableTwoFactor: phttp.UseCaseFunc[usersapi.TwoFactorDisableInput, struct{}](func(ctx context.Context, cmd usersapi.TwoFactorDisableInput) (struct{}, error) {
			return struct{}{}, svc.DisableTwoFactor(ctx, cmd)
		}),
		challengeStatus: phttp.UseCaseFunc[usersapi.ChallengeStatusInput, login.Output](func(ctx context.Context, cmd usersapi.ChallengeStatusInput) (login.Output, error) {
			return svc.ChallengeStatus(ctx, cmd)
		}),
		challengeVerifyTOTP: phttp.UseCaseFunc[usersapi.ChallengeVerifyTOTPInput, login.Output](func(ctx context.Context, cmd usersapi.ChallengeVerifyTOTPInput) (login.Output, error) {
			return svc.VerifyChallengeTOTP(ctx, cmd)
		}),
		challengeResendEmail: phttp.UseCaseFunc[usersapi.ChallengeResendEmailInput, login.Output](func(ctx context.Context, cmd usersapi.ChallengeResendEmailInput) (login.Output, error) {
			return svc.ResendChallengeEmail(ctx, cmd)
		}),
		challengeConfirmEmail: phttp.UseCaseFunc[usersapi.ChallengeConfirmEmailInput, login.Output](func(ctx context.Context, cmd usersapi.ChallengeConfirmEmailInput) (login.Output, error) {
			return svc.ConfirmChallengeEmail(ctx, cmd)
		}),
		getMe: phttp.UseCaseFunc[usersapi.GetProfileInput, profile.Output](func(ctx context.Context, cmd usersapi.GetProfileInput) (profile.Output, error) {
			return svc.GetMe(ctx, cmd)
		}),
		update: phttp.UseCaseFunc[usersapi.UpdateProfileInput, profile.Output](func(ctx context.Context, cmd usersapi.UpdateProfileInput) (profile.Output, error) {
			return svc.UpdateProfile(ctx, cmd)
		}),
		changePassword: phttp.UseCaseFunc[usersapi.ChangePasswordInput, struct{}](func(ctx context.Context, cmd usersapi.ChangePasswordInput) (struct{}, error) {
			return struct{}{}, svc.ChangePassword(ctx, cmd)
		}),
		link: phttp.UseCaseFunc[usersapi.LinkProviderInput, link.Output](func(ctx context.Context, cmd usersapi.LinkProviderInput) (link.Output, error) {
			return svc.LinkProvider(ctx, cmd)
		}),
		listSessions: phttp.UseCaseFunc[usersapi.ListSessionsInput, usersapi.SessionsOutput](func(ctx context.Context, cmd usersapi.ListSessionsInput) (usersapi.SessionsOutput, error) {
			return svc.ListSessions(ctx, cmd)
		}),
		revokeSession: phttp.UseCaseFunc[usersapi.RevokeSessionInput, struct{}](func(ctx context.Context, cmd usersapi.RevokeSessionInput) (struct{}, error) {
			return struct{}{}, svc.RevokeSession(ctx, cmd)
		}),
		revokeOtherSession: phttp.UseCaseFunc[usersapi.RevokeOtherSessionsInput, struct{}](func(ctx context.Context, cmd usersapi.RevokeOtherSessionsInput) (struct{}, error) {
			return struct{}{}, svc.RevokeOtherSessions(ctx, cmd)
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
			Email:       out.Email,
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
		OTP:      req.OTP,
	})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}
	writeAuthResponse(w, out)
}

func (h *Handler) ChallengeStatus(w http.ResponseWriter, r *http.Request) {
	var req dto.ChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}

	out, err := phttp.HandleUseCase(h.middleware, r, h.challengeStatus, usersapi.ChallengeStatusInput{ChallengeID: req.ChallengeID})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}
	writeAuthResponse(w, out)
}

func (h *Handler) VerifyChallengeTOTP(w http.ResponseWriter, r *http.Request) {
	var req dto.ChallengeTOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}
	out, err := phttp.HandleUseCase(h.middleware, r, h.challengeVerifyTOTP, usersapi.ChallengeVerifyTOTPInput{ChallengeID: req.ChallengeID, Code: req.Code})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}
	writeAuthResponse(w, out)
}

func (h *Handler) ResendChallengeEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.ChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}
	out, err := phttp.HandleUseCase(h.middleware, r, h.challengeResendEmail, usersapi.ChallengeResendEmailInput{ChallengeID: req.ChallengeID})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}
	writeAuthResponse(w, out)
}

func (h *Handler) ConfirmChallengeEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.ChallengeConfirmEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}
	out, err := phttp.HandleUseCase(h.middleware, r, h.challengeConfirmEmail, usersapi.ChallengeConfirmEmailInput{ChallengeID: req.ChallengeID, Token: req.Token})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}
	writeAuthResponse(w, out)
}

func toChallengeDTO(info *login.ChallengeInfo, status string) *dto.ChallengeResponse {
	if info == nil {
		return nil
	}
	return &dto.ChallengeResponse{
		Status:         status,
		ChallengeID:    info.ID,
		Type:           info.Type,
		RequiredSteps:  info.RequiredSteps,
		CompletedSteps: info.CompletedSteps,
		ExpiresIn:      info.ExpiresIn,
		MaskedEmail:    info.MaskedEmail,
		AttemptsLeft:   info.AttemptsLeft,
		LockUntil:      info.LockUntil,
	}
}

func toProfileDTO(out profile.Output) dto.UserProfileResponse {
	return dto.UserProfileResponse{
		UserID:      out.UserID,
		Email:       out.Email,
		FirstName:   out.FirstName,
		LastName:    out.LastName,
		MiddleName:  out.MiddleName,
		DisplayName: out.DisplayName,
		AvatarURL:   out.AvatarURL,
		LoginSettings: dto.LoginSettingsResponse{
			TwoFactorEnabled: out.LoginSettings.TwoFactorEnabled,
			EmailVerified:    out.LoginSettings.EmailVerified,
		},
	}
}

func writeAuthResponse(w http.ResponseWriter, out login.Output) {
	if out.Status == "challenge_required" && out.Challenge != nil {
		phttp.WriteJSON(w, http.StatusOK, toChallengeDTO(out.Challenge, out.Status))
		return
	}

	phttp.WriteJSON(w, http.StatusOK, dto.LoginResponse{
		UserProfileResponse: dto.UserProfileResponse{
			UserID:      out.UserID,
			Email:       out.Email,
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
		Challenge: toChallengeDTO(out.Challenge, out.Status),
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
			Email:       out.Email,
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

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.SocialIDTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}

	out, err := phttp.HandleUseCase(h.middleware, r, h.google, usersapi.GoogleLoginInput{IDToken: req.IDToken})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}
	writeAuthResponse(w, out)
}

func (h *Handler) AppleLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.SocialIDTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}

	out, err := phttp.HandleUseCase(h.middleware, r, h.apple, usersapi.AppleLoginInput{IDToken: req.IDToken})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}
	writeAuthResponse(w, out)
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

func (h *Handler) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.ConfirmEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}

	out, err := phttp.HandleUseCase(h.middleware, r, h.confirmEmail, usersapi.ConfirmEmailInput{
		Email: req.Email,
		Code:  req.Code,
	})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}

	phttp.WriteJSON(w, http.StatusOK, dto.LoginResponse{
		UserProfileResponse: dto.UserProfileResponse{
			UserID:      out.UserID,
			Email:       out.Email,
			FirstName:   out.FirstName,
			LastName:    out.LastName,
			MiddleName:  out.MiddleName,
			DisplayName: out.DisplayName,
			AvatarURL:   out.AvatarURL,
		},
		TokensResponse: dto.TokensResponse{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken},
	})
}

func (h *Handler) RequestEmailConfirmation(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}
	if _, err := phttp.HandleUseCase(h.middleware, r, h.requestConfirm, usersapi.RequestEmailInput{Email: req.Email}); err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}
	if _, err := phttp.HandleUseCase(h.middleware, r, h.requestPasswordReset, usersapi.RequestPasswordResetInput{Email: req.Email}); err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}
	if _, err := phttp.HandleUseCase(h.middleware, r, h.resetPassword, usersapi.ResetPasswordInput{Email: req.Email, Token: req.Token, Code: req.Code, NewPassword: req.Password}); err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

	phttp.WriteJSON(w, http.StatusOK, toProfileDTO(out))
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

	phttp.WriteJSON(w, http.StatusOK, toProfileDTO(out))
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	uid, ok := httpctx.UserIDFromContext(r.Context())
	if !ok {
		phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}

	if _, err := phttp.HandleUseCase(h.middleware, r, h.changePassword, usersapi.ChangePasswordInput{UserID: uid, CurrentPassword: req.CurrentPassword, NewPassword: req.NewPassword}); err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	uid, ok := httpctx.UserIDFromContext(r.Context())
	if !ok {
		phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	current := r.URL.Query().Get("current_refresh_token")
	if current == "" {
		current = r.Header.Get("X-Refresh-Token")
	}

	out, err := phttp.HandleUseCase(h.middleware, r, h.listSessions, usersapi.ListSessionsInput{
		UserID:              uid,
		CurrentRefreshToken: current,
	})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}

	resp := dto.SessionsResponse{Sessions: make([]dto.SessionResponse, 0, len(out.Sessions))}
	for _, s := range out.Sessions {
		resp.Sessions = append(resp.Sessions, dto.SessionResponse{
			ID:        s.ID,
			UserAgent: s.UserAgent,
			IP:        s.IP,
			CreatedAt: s.CreatedAt,
			ExpiresAt: s.ExpiresAt,
			RevokedAt: s.RevokedAt,
			Current:   s.Current,
		})
	}

	phttp.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	uid, ok := httpctx.UserIDFromContext(r.Context())
	if !ok {
		phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	var req dto.RevokeSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}
	if req.SessionID == "" {
		phttp.WriteError(w, http.StatusBadRequest, "validation_error", "session_id is required")
		return
	}

	if _, err := phttp.HandleUseCase(h.middleware, r, h.revokeSession, usersapi.RevokeSessionInput{UserID: uid, SessionID: req.SessionID}); err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RevokeOtherSessions(w http.ResponseWriter, r *http.Request) {
	uid, ok := httpctx.UserIDFromContext(r.Context())
	if !ok {
		phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}

	var req dto.RevokeOtherSessionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}
	if req.CurrentRefreshToken == "" {
		req.CurrentRefreshToken = r.Header.Get("X-Refresh-Token")
	}
	if req.CurrentRefreshToken == "" {
		phttp.WriteError(w, http.StatusBadRequest, "validation_error", "current_refresh_token is required")
		return
	}

	if _, err := phttp.HandleUseCase(h.middleware, r, h.revokeOtherSession, usersapi.RevokeOtherSessionsInput{
		UserID:              uid,
		CurrentRefreshToken: req.CurrentRefreshToken,
	}); err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SetupTwoFactor(w http.ResponseWriter, r *http.Request) {
	uid, ok := httpctx.UserIDFromContext(r.Context())
	if !ok {
		phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}
	out, err := phttp.HandleUseCase(h.middleware, r, h.setupTwoFactor, usersapi.TwoFactorSetupInput{UserID: uid})
	if err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}
	phttp.WriteJSON(w, http.StatusOK, dto.TwoFactorSetupResponse{Secret: out.Secret, URI: out.ProvisioningQR})
}

func (h *Handler) ConfirmTwoFactor(w http.ResponseWriter, r *http.Request) {
	uid, ok := httpctx.UserIDFromContext(r.Context())
	if !ok {
		phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}
	var req dto.TwoFactorCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}
	if _, err := phttp.HandleUseCase(h.middleware, r, h.confirmTwoFactor, usersapi.TwoFactorConfirmInput{UserID: uid, Code: req.Code}); err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DisableTwoFactor(w http.ResponseWriter, r *http.Request) {
	uid, ok := httpctx.UserIDFromContext(r.Context())
	if !ok {
		phttp.WriteError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
		return
	}
	var req dto.TwoFactorCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		phttp.WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON")
		return
	}
	if _, err := phttp.HandleUseCase(h.middleware, r, h.disableTwoFactor, usersapi.TwoFactorDisableInput{UserID: uid, Code: req.Code}); err != nil {
		status, code, msg := mapError(err)
		phttp.WriteError(w, status, code, msg)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func mapError(err error) (status int, code string, message string) {
	if errors.Is(err, domain.ErrInvalidEmail) || errors.Is(err, domain.ErrWeakPassword) || errors.Is(err, domain.ErrInvalidDisplayName) || errors.Is(err, domain.ErrInvalidAvatarURL) {
		return http.StatusBadRequest, "validation_error", "Validation error"
	}
	if errors.Is(err, domain.ErrEmailAlreadyUsed) || errors.Is(err, domain.ErrIdentityAlreadyLinked) {
		return http.StatusConflict, "conflict", "Conflict"
	}
	if errors.Is(err, domain.ErrInvalidCredentials) || errors.Is(err, domain.ErrRefreshTokenInvalid) {
		return http.StatusUnauthorized, "unauthorized", "Unauthorized"
	}
	if errors.Is(err, domain.ErrEmailNotVerified) {
		return http.StatusForbidden, "email_not_verified", "Email not verified"
	}
	if errors.Is(err, domain.ErrTwoFactorRequired) || errors.Is(err, domain.ErrInvalidTwoFactor) {
		return http.StatusUnauthorized, "two_factor_required", "Two-factor verification required"
	}
	if errors.Is(err, domain.ErrTwoFactorAlreadyEnabled) {
		return http.StatusConflict, "two_factor_already_enabled", "Two-factor is already enabled"
	}
	if errors.Is(err, domain.ErrTooManyRequests) {
		return http.StatusTooManyRequests, "too_many_requests", "Too many requests"
	}
	if errors.Is(err, domain.ErrUnauthorized) {
		return http.StatusUnauthorized, "unauthorized", "Unauthorized"
	}
	if errors.Is(err, common.ErrInternal) {
		return http.StatusInternalServerError, "internal_error", "Internal server error"
	}
	return http.StatusInternalServerError, "internal_error", "Internal server error"
}
