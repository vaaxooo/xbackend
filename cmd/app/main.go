package main

import (
	"log"

	"github.com/vaaxooo/xbackend/internal/app"
	"github.com/vaaxooo/xbackend/internal/platform/config"
	pdb "github.com/vaaxooo/xbackend/internal/platform/db"
	plog "github.com/vaaxooo/xbackend/internal/platform/log"
)

func main() {
	container, cleanup, err := buildApplication()
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	if err := container.Run(); err != nil {
		log.Fatal(err)
	}
}

func buildApplication() (*app.Container, func(), error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, func() {}, err
	}

	logger := plog.New(cfg.App.Env)

	db, err := pdb.OpenPostgres(
		cfg.DB.DSN,
		cfg.DB.MaxOpenConnects,
		cfg.DB.MaxIdleConnects,
		cfg.DB.ConnMaxLife,
	)
	if err != nil {
		return nil, func() {}, err
	}

	container, err := app.NewContainer(app.Deps{
		Config: cfg,
		Logger: logger,
		DB:     db,
	})
	if err != nil {
		_ = db.Close()
		return nil, func() {}, err
	}

	cleanup := func() {
		_ = db.Close()
	}

	return container, cleanup, nil
}
