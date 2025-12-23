package profile

type GetInput struct {
	UserID string
}
type UpdateInput struct {
	UserID      string
	FirstName   *string
	LastName    *string
	MiddleName  *string
	DisplayName *string
	AvatarURL   *string
}

type LoginSettings struct {
	TwoFactorEnabled bool
	EmailVerified    bool
}
type Output struct {
	UserID        string
	Email         string
	FirstName     string
	LastName      string
	MiddleName    string
	DisplayName   string
	AvatarURL     string
	LoginSettings LoginSettings
}
