package users

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/vaaxooo/xbackend/internal/modules/users/public"
	phttp "github.com/vaaxooo/xbackend/internal/platform/http"
	"github.com/vaaxooo/xbackend/internal/platform/http/users/middleware"
	pmiddleware "github.com/vaaxooo/xbackend/internal/platform/middleware"
)

func RegisterV1(r chi.Router, svc public.Service, auth public.AuthPort) {
	h := NewHandler(svc, phttp.UseCaseMiddleware{Timeout: 30 * time.Second})

	r.Route("/auth", func(r chi.Router) {
		// Auth endpoints are brute-force targets.
		r.With(pmiddleware.RateLimit(20, time.Minute)).Post("/register", h.Register)
		r.With(pmiddleware.RateLimit(10, time.Minute)).Post("/login", h.Login)
		r.With(pmiddleware.RateLimit(10, time.Minute)).Post("/telegram", h.TelegramLogin)
		r.With(pmiddleware.RateLimit(20, time.Minute)).Post("/refresh", h.Refresh)
		r.With(pmiddleware.RateLimit(20, time.Minute)).Post("/confirm", h.ConfirmEmail)
		r.With(pmiddleware.RateLimit(10, time.Minute)).Post("/confirm/request", h.RequestEmailConfirmation)
		r.With(pmiddleware.RateLimit(10, time.Minute)).Post("/password/reset", h.RequestPasswordReset)
		r.With(pmiddleware.RateLimit(10, time.Minute)).Post("/password/confirm", h.ResetPassword)
		r.With(pmiddleware.RateLimit(20, time.Minute)).Post("/challenge/status", h.ChallengeStatus)
		r.With(pmiddleware.RateLimit(20, time.Minute)).Post("/challenge/verify-totp", h.VerifyChallengeTOTP)
		r.With(pmiddleware.RateLimit(20, time.Minute)).Post("/challenge/resend-email", h.ResendChallengeEmail)
		r.With(pmiddleware.RateLimit(20, time.Minute)).Post("/challenge/confirm-email", h.ConfirmChallengeEmail)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireJWT(auth))
			r.Post("/link", h.LinkProvider)
			r.Post("/password/change", h.ChangePassword)
			r.Post("/2fa/setup", h.SetupTwoFactor)
			r.Post("/2fa/confirm", h.ConfirmTwoFactor)
			r.Post("/2fa/disable", h.DisableTwoFactor)
			r.Get("/sessions", h.ListSessions)
			r.Post("/sessions/revoke", h.RevokeSession)
			r.Post("/sessions/revoke-others", h.RevokeOtherSessions)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireJWT(auth))
		r.Get("/me", h.GetMe)
		r.Patch("/me", h.UpdateProfile)
	})

}
