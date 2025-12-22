package http

import (
	middlewares "github.com/vaaxooo/xbackend/internal/platform/middleware"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	plog "github.com/vaaxooo/xbackend/internal/platform/log"
)

type RouterDeps struct {
	Logger             plog.Logger
	Timeout            time.Duration
	CORSAllowedOrigins []string
}

func NewRouter(deps RouterDeps, registerAPIV1 func(r chi.Router)) http.Handler {
	r := chi.NewRouter()

	// def middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.BodyLimit(1 << 20))
	r.Use(middleware.Timeout(deps.Timeout))

	if len(deps.CORSAllowedOrigins) > 0 {
		r.Use(middlewares.CORS(deps.CORSAllowedOrigins))
	}

	// access log
	r.Use(middlewares.AccessLog(deps.Logger))

	// healthcheck
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// versioned API
	r.Route("/api/v1", func(api chi.Router) {
		if registerAPIV1 != nil {
			registerAPIV1(api)
		}
	})

	return r
}
