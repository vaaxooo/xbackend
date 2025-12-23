package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "backend"),
			Env:  getEnv("APP_ENV", "dev"),
		},
		HTTP: HTTPConfig{
			Addr:               getEnv("HTTP_ADDR", ":8080"),
			ReadTimeout:        getDuration("HTTP_READ_TIMEOUT", 5*time.Second),
			ReadHeaderTimeout:  getDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
			WriteTimeout:       getDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:        getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			MaxHeaderBytes:     getInt("HTTP_MAX_HEADER_BYTES", 0),
			CORSAllowedOrigins: getStringSlice("CORS_ALLOWED_ORIGINS"),
		},
		DB: DBConfig{
			Driver:          getEnv("DB_DRIVER", "postgres"),
			DSN:             getEnv("DB_DSN", ""),
			MaxOpenConnects: getInt("DB_MAX_OPEN_CONNECTS", 25),
			MaxIdleConnects: getInt("DB_MAX_IDLE_CONNECTS", 25),
			ConnMaxLife:     getDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Auth: AuthConfig{
			JWTSecret:                getEnv("AUTH_JWT_SECRET", ""),
			AccessTTL:                getDuration("AUTH_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:               getDuration("AUTH_REFRESH_TTL", 30*24*time.Hour),
			RefreshRetentionTTL:      getDuration("AUTH_REFRESH_RETENTION_TTL", 90*24*time.Hour),
			RequireEmailConfirmation: getBool("AUTH_REQUIRE_EMAIL_CONFIRMATION", false),
			VerificationTTL:          getDuration("AUTH_VERIFICATION_TTL", 15*time.Minute),
			PasswordResetTTL:         getDuration("AUTH_PASSWORD_RESET_TTL", 15*time.Minute),
			TwoFactorIssuer:          getEnv("AUTH_TWO_FACTOR_ISSUER", "xbackend"),
		},
		Telegram: TelegramConfig{
			BotToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
			InitDataTTL: getDuration("TELEGRAM_INIT_DATA_TTL", 24*time.Hour),
		},
		Google: GoogleConfig{
			ClientID: getEnv("GOOGLE_CLIENT_ID", ""),
			JWKSURL:  getEnv("GOOGLE_JWKS_URL", "https://www.googleapis.com/oauth2/v3/certs"),
		},
		Apple: AppleConfig{
			ClientID: getEnv("APPLE_CLIENT_ID", ""),
			JWKSURL:  getEnv("APPLE_JWKS_URL", "https://appleid.apple.com/auth/keys"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
			UseTLS:   getBool("SMTP_USE_TLS", true),
			Timeout:  getDuration("SMTP_TIMEOUT", 10*time.Second),
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

func getBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return parsed
}

func getStringSlice(key string) []string {
	raw := os.Getenv(key)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
