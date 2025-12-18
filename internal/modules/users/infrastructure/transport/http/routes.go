package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/transport/middleware"
	phttp "github.com/vaaxooo/xbackend/internal/platform/middleware"

	usersapp "github.com/vaaxooo/xbackend/internal/modules/users/application"

	"time"
)

func RegisterV1(r chi.Router, svc usersapp.Service, tp middleware.TokenParser) {
	h := NewHandler(svc)

	r.Route("/auth", func(r chi.Router) {
		// Auth endpoints are brute-force targets.
		r.With(phttp.RateLimit(20, time.Minute)).Post("/register", h.Register)
		r.With(phttp.RateLimit(10, time.Minute)).Post("/login", h.Login)
		r.With(phttp.RateLimit(20, time.Minute)).Post("/refresh", h.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireJWT(tp))
			r.Post("/link", h.LinkProvider)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireJWT(tp))
		r.Get("/me", h.GetMe)
		r.Patch("/me", h.UpdateProfile)
	})

}
