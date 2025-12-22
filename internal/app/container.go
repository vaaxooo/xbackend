package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/vaaxooo/xbackend/internal/modules/users/application/common"
	usersevents "github.com/vaaxooo/xbackend/internal/modules/users/infrastructure/events"
	pconfig "github.com/vaaxooo/xbackend/internal/platform/config"
	phttp "github.com/vaaxooo/xbackend/internal/platform/http"
	plog "github.com/vaaxooo/xbackend/internal/platform/log"
	"github.com/vaaxooo/xbackend/internal/platform/outbox"
)

type Container struct {
	cfg    *pconfig.Config
	logger plog.Logger

	db      *sql.DB
	modules *Modules
	server  *phttp.Server
	worker  *outbox.Worker
}

type Deps struct {
	Config *pconfig.Config
	Logger plog.Logger
	DB     *sql.DB
}

func NewContainer(deps Deps) (*Container, error) {
	// Initialize all modules (bounded contexts).
	mods, err := InitModules(
		ModuleDeps{
			DB:     deps.DB,
			Logger: deps.Logger,
		},
		ModulesConfig{Users: UsersConfig(deps.Config)},
	)
	if err != nil {
		return nil, err
	}

	// Build HTTP router with versioned API registrations.
        handler := phttp.NewRouter(
                phttp.RouterDeps{
                        Logger:            deps.Logger,
                        Timeout:           30 * time.Second,
                        CORSAllowedOrigins: deps.Config.HTTP.CORSAllowedOrigins,
                },
                func(r chi.Router) { RegisterAPIV1(r, mods) },
        )

	server := phttp.NewServer(
		phttp.ServerConfig{
			Addr:              deps.Config.HTTP.Addr,
			ReadTimeout:       deps.Config.HTTP.ReadTimeout,
			ReadHeaderTimeout: deps.Config.HTTP.ReadHeaderTimeout,
			WriteTimeout:      deps.Config.HTTP.WriteTimeout,
			IdleTimeout:       deps.Config.HTTP.IdleTimeout,
			MaxHeaderBytes:    deps.Config.HTTP.MaxHeaderBytes,
		},
		handler,
	)

	var domainPublisher common.DomainEventPublisher = usersevents.NewLoggerDomainPublisher(deps.Logger)
	if deps.Config.SMTP.Host != "" && deps.Config.SMTP.From != "" {
		mailer := usersevents.NewSMTPMailer(
			deps.Config.SMTP.Host,
			deps.Config.SMTP.Port,
			deps.Config.SMTP.Username,
			deps.Config.SMTP.Password,
			deps.Config.SMTP.From,
			deps.Config.SMTP.UseTLS,
			deps.Config.SMTP.Timeout,
		)
		domainPublisher = usersevents.NewMultiDomainPublisher(
			usersevents.NewOutboxEmailPublisher(mailer, deps.Logger),
			domainPublisher,
		)
	}

	worker := outbox.NewWorker(mods.Users.Outbox, domainPublisher, deps.Logger, outbox.Config{})

	return &Container{
		cfg:     deps.Config,
		logger:  deps.Logger,
		db:      deps.DB,
		modules: mods,
		server:  server,
		worker:  worker,
	}, nil
}

func (c *Container) Run() error {
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

	if c.worker != nil {
		go c.worker.Run(ctx)
	}

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
