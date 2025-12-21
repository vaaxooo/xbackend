package dto

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
        Email    string `json:"email"`
        Password string `json:"password"`
        OTP       string `json:"otp_code"`
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

type TwoFactorSetupResponse struct {
        Secret string `json:"secret"`
        URI    string `json:"uri"`
}

type TwoFactorCodeRequest struct {
        Code string `json:"code"`
}

type UserProfileResponse struct {
	UserID      string `json:"user_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	MiddleName  string `json:"middle_name"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type LoginResponse struct {
	UserProfileResponse
	TokensResponse
}
