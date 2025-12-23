package dto

import "time"

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	OTP      string `json:"otp_code"`
}

type TelegramLoginRequest struct {
	InitData string `json:"init_data"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ConfirmEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type PasswordResetRequest struct {
	Email string `json:"email"`
}

type PasswordResetConfirmRequest struct {
	Email    string `json:"email"`
	Code     string `json:"code"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type TwoFactorSetupResponse struct {
	Secret string `json:"secret"`
	URI    string `json:"uri"`
}

type TwoFactorCodeRequest struct {
	Code string `json:"code"`
}

type UserProfileResponse struct {
	UserID        string                `json:"user_id"`
	Email         string                `json:"email"`
	FirstName     string                `json:"first_name"`
	LastName      string                `json:"last_name"`
	MiddleName    string                `json:"middle_name"`
	DisplayName   string                `json:"display_name"`
	AvatarURL     string                `json:"avatar_url"`
	LoginSettings LoginSettingsResponse `json:"login_settings"`
}

type LoginSettingsResponse struct {
	TwoFactorEnabled bool `json:"two_factor_enabled"`
	EmailVerified    bool `json:"email_verified"`
}

type LoginResponse struct {
	UserProfileResponse
	TokensResponse
	Challenge *ChallengeResponse `json:"challenge,omitempty"`
}

type ChallengeResponse struct {
	Status         string     `json:"status"`
	ChallengeID    string     `json:"challenge_id"`
	Type           string     `json:"challenge_type"`
	RequiredSteps  []string   `json:"required_steps"`
	CompletedSteps []string   `json:"completed_steps"`
	ExpiresIn      int64      `json:"expires_in"`
	MaskedEmail    string     `json:"masked_email,omitempty"`
	AttemptsLeft   int        `json:"attempts_left,omitempty"`
	LockUntil      *time.Time `json:"lock_until,omitempty"`
}

type ChallengeRequest struct {
	ChallengeID string `json:"challenge_id"`
}

type ChallengeTOTPRequest struct {
	ChallengeID string `json:"challenge_id"`
	Code        string `json:"otp_code"`
}

type ChallengeConfirmEmailRequest struct {
	ChallengeID string `json:"challenge_id"`
	Token       string `json:"token"`
}
