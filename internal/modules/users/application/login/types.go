package login

import "time"

type Input struct {
	Email    string
	Password string
	OTP      string
}

type ChallengeInfo struct {
	ID             string
	Type           string
	RequiredSteps  []string
	CompletedSteps []string
	Status         string
	ExpiresIn      int64
	AttemptsLeft   int
	LockUntil      *time.Time
	MaskedEmail    string
}
type Output struct {
	UserID       string
	Email        string
	FirstName    string
	LastName     string
	MiddleName   string
	DisplayName  string
	AvatarURL    string
	AccessToken  string
	RefreshToken string

	Status    string
	Challenge *ChallengeInfo
}
