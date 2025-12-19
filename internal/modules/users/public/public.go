package public

import (
	"time"

	"github.com/vaaxooo/xbackend/internal/modules/users/application"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/link"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/login"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/profile"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/refresh"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/register"
	"github.com/vaaxooo/xbackend/internal/modules/users/application/telegram"
)

type Service = application.Service

type Config struct {
	Auth     AuthConfig
	Telegram TelegramConfig
}

type AuthConfig struct {
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type TelegramConfig struct {
	BotToken    string
	InitDataTTL time.Duration
}

type AuthPort interface {
	Issue(userID string, ttl time.Duration) (string, error)
	Verify(token string) (AuthContext, error)
}

type AuthContext struct {
	UserID string
	Roles  []string
}

// Re-export DTOs and commands used by transports.
type RegisterInput = register.Input
type LoginInput = login.Input
type LoginOutput = login.Output
type RefreshInput = refresh.Input
type RefreshOutput = refresh.Output
type TelegramLoginInput = telegram.Input
type GetProfileInput = profile.GetInput
type UpdateProfileInput = profile.UpdateInput
type ProfileOutput = profile.Output
type LinkProviderInput = link.Input
type LinkProviderOutput = link.Output
