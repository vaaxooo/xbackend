package login

type Input struct {
	Email    string
	Password string
	OTP      string
}
type Output struct {
	UserID       string
	FirstName    string
	LastName     string
	MiddleName   string
	DisplayName  string
	AvatarURL    string
	AccessToken  string
	RefreshToken string
}
