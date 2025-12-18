package app

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pconfig "github.com/vaaxooo/xbackend/internal/platform/config"
	pdb "github.com/vaaxooo/xbackend/internal/platform/db"
	phttp "github.com/vaaxooo/xbackend/internal/platform/http"
	plog "github.com/vaaxooo/xbackend/internal/platform/log"
)

type Container struct {
	cfg    *pconfig.Config
	logger plog.Logger

	db      *sql.DB
	modules *Modules
	server  *phttp.Server
}

func NewContainer(cfg *pconfig.Config) *Container {
	return &Container{cfg: cfg}
}

func (c *Container) Run() error {
	c.logger = plog.New(c.cfg.App.Env)

	// Open infrastructure dependencies first (fail-fast on startup).
	db, err := pdb.OpenPostgres(
		c.cfg.DB.DSN,
		c.cfg.DB.MaxOpenConnects,
		c.cfg.DB.MaxIdleConnects,
		c.cfg.DB.ConnMaxLife,
	)
	if err != nil {
		c.logger.Error(context.Background(), "failed to open postgres", err)
		return err
	}
	c.db = db
	defer func() { _ = c.db.Close() }()

	// Initialize all modules (bounded contexts).
	mods, err := InitModules(ModuleDeps{
		DB:     c.db,
		Logger: c.logger,

		AuthJWTSecret:  c.cfg.Auth.JWTSecret,
		AuthAccessTTL:  c.cfg.Auth.AccessTTL,
		AuthRefreshTTL: c.cfg.Auth.RefreshTTL,
	})
	if err != nil {
		c.logger.Error(context.Background(), "failed to init modules", err)
		return err
	}
	c.modules = mods

	// Build HTTP router with versioned API registrations.
	handler := phttp.NewRouter(
		phttp.RouterDeps{
			Logger:  c.logger,
			Timeout: 30 * time.Second,
		},
		func(r chi.Router) { RegisterAPIV1(r, c.modules) },
	)

	c.server = phttp.NewServer(
		phttp.ServerConfig{
			Addr:              c.cfg.HTTP.Addr,
			ReadTimeout:       c.cfg.HTTP.ReadTimeout,
			ReadHeaderTimeout: c.cfg.HTTP.ReadHeaderTimeout,
			WriteTimeout:      c.cfg.HTTP.WriteTimeout,
			IdleTimeout:       c.cfg.HTTP.IdleTimeout,
			MaxHeaderBytes:    c.cfg.HTTP.MaxHeaderBytes,
		},
		handler,
	)

	// Graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		c.logger.Info(context.Background(), "http server starting", "addr", c.cfg.HTTP.Addr)
		if err := c.server.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		c.logger.Info(context.Background(), "shutdown signal received")
	case err := <-errCh:
		c.logger.Error(context.Background(), "http server error", err)
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.server.Shutdown(shutdownCtx); err != nil {
		c.logger.Error(context.Background(), "http server shutdown failed", err)
		return err
	}

	c.logger.Info(context.Background(), "server stopped gracefully")
	return nil
}
