package http

import (
	"context"
	"net"
	"net/http"
	"time"
)

type Server struct {
	srv *http.Server
}

type ServerConfig struct {
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
}

func NewServer(cfg ServerConfig, handler http.Handler) *Server {
	// Reasonable defaults for production safety.
	if cfg.ReadHeaderTimeout == 0 {
		cfg.ReadHeaderTimeout = 5 * time.Second
	}
	if cfg.MaxHeaderBytes == 0 {
		cfg.MaxHeaderBytes = 1 << 20 // 1 MB
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,

		// BaseContext sets the base context for incoming requests.
		// It helps to tie request contexts to a predictable root context.
		BaseContext: func(_ net.Listener) context.Context {
			return context.Background()
		},
	}

	return &Server{srv: srv}
}

func (s *Server) Run() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
