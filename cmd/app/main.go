package main

import (
	"log"

	"github.com/vaaxooo/xbackend/internal/app"
	"github.com/vaaxooo/xbackend/internal/platform/config"
)

func main() {
	// Load application configuration from environment.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Build application container (composition root).
	container := app.NewContainer(cfg)

	// Run application.
	if err := container.Run(); err != nil {
		log.Fatal(err)
	}
}
