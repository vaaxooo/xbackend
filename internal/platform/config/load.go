package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "backend"),
			Env:  getEnv("APP_ENV", "dev"),
		},
		HTTP: HTTPConfig{
			Addr:              getEnv("HTTP_ADDR", ":8080"),
			ReadTimeout:       getDuration("HTTP_READ_TIMEOUT", 5*time.Second),
			ReadHeaderTimeout: getDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
			WriteTimeout:      getDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:       getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			MaxHeaderBytes:    getInt("HTTP_MAX_HEADER_BYTES", 0),
		},
		DB: DBConfig{
			Driver:          getEnv("DB_DRIVER", "postgres"),
			DSN:             getEnv("DB_DSN", ""),
			MaxOpenConnects: getInt("DB_MAX_OPEN_CONNECTS", 25),
			MaxIdleConnects: getInt("DB_MAX_IDLE_CONNECTS", 25),
			ConnMaxLife:     getDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Auth: AuthConfig{
			JWTSecret:  getEnv("AUTH_JWT_SECRET", ""),
			AccessTTL:  getDuration("AUTH_ACCESS_TTL", 15*time.Minute),
			RefreshTTL: getDuration("AUTH_REFRESH_TTL", 30*24*time.Hour),
		},
		Telegram: TelegramConfig{
			BotToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
			InitDataTTL: getDuration("TELEGRAM_INIT_DATA_TTL", 24*time.Hour),
		},
	}

	if cfg.DB.DSN == "" {
		return nil, fmt.Errorf("DB_DSN is required")
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
