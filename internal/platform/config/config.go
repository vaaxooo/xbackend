package config

import "time"

type Config struct {
	App  AppConfig
	HTTP HTTPConfig
	DB   DBConfig
	Auth AuthConfig
}

type AppConfig struct {
	Name string
	Env  string // dev | prod
}

type HTTPConfig struct {
	Addr              string // ":8080"
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
}

type DBConfig struct {
	Driver string // postgres
	DSN    string

	MaxOpenConnects int
	MaxIdleConnects int
	ConnMaxLife     time.Duration
}

type AuthConfig struct {
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}
