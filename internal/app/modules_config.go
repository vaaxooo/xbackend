package app

import (
	userspublic "github.com/vaaxooo/xbackend/internal/modules/users/public"
	pconfig "github.com/vaaxooo/xbackend/internal/platform/config"
)

// UsersConfig builds the configuration for the Users bounded context
// from the global application config.
func UsersConfig(cfg *pconfig.Config) userspublic.Config {
	return userspublic.Config{
		Auth: userspublic.AuthConfig{
			JWTSecret:                cfg.Auth.JWTSecret,
			AccessTTL:                cfg.Auth.AccessTTL,
			RefreshTTL:               cfg.Auth.RefreshTTL,
			RequireEmailConfirmation: cfg.Auth.RequireEmailConfirmation,
			VerificationTTL:          cfg.Auth.VerificationTTL,
			PasswordResetTTL:         cfg.Auth.PasswordResetTTL,
			TwoFactorIssuer:          cfg.Auth.TwoFactorIssuer,
		},
		Telegram: userspublic.TelegramConfig{
			BotToken:    cfg.Telegram.BotToken,
			InitDataTTL: cfg.Telegram.InitDataTTL,
		},
	}
}
