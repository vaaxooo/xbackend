package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChallengeStatus string

type ChallengeStep string

const (
	ChallengeStatusPending   ChallengeStatus = "pending"
	ChallengeStatusCompleted ChallengeStatus = "completed"
	ChallengeStatusBlocked   ChallengeStatus = "blocked"
	ChallengeStatusExpired   ChallengeStatus = "expired"

	ChallengeStepTOTP              ChallengeStep = "totp"
	ChallengeStepEmailVerification ChallengeStep = "email_verification"
	ChallengeStepAccountBlocked    ChallengeStep = "account_blocked"
	ChallengeStepCaptcha           ChallengeStep = "captcha"
)

type Challenge struct {
	ID                 string
	UserID             UserID
	Type               string
	RequiredSteps      []ChallengeStep
	CompletedSteps     []ChallengeStep
	Status             ChallengeStatus
	ExpiresAt          time.Time
	SessionFingerprint string
	AttemptsLeft       int
	LockUntil          *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewChallenge(userID UserID, challengeType string, required []ChallengeStep, expiresAt time.Time) Challenge {
	return Challenge{
		ID:             uuid.NewString(),
		UserID:         userID,
		Type:           challengeType,
		RequiredSteps:  required,
		CompletedSteps: []ChallengeStep{},
		Status:         ChallengeStatusPending,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
}

func (c Challenge) IsExpired(now time.Time) bool {
	return now.After(c.ExpiresAt)
}

func (c Challenge) NeedsStep(step ChallengeStep) bool {
	for _, required := range c.RequiredSteps {
		if required == step {
			for _, done := range c.CompletedSteps {
				if done == step {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (c Challenge) WithCompleted(step ChallengeStep, now time.Time) Challenge {
	if !c.NeedsStep(step) {
		return c
	}
	c.CompletedSteps = append(c.CompletedSteps, step)
	c.UpdatedAt = now
	if len(c.CompletedSteps) == len(c.RequiredSteps) {
		c.Status = ChallengeStatusCompleted
	}
	return c
}

func (c Challenge) WithStatus(status ChallengeStatus, now time.Time) Challenge {
	c.Status = status
	c.UpdatedAt = now
	return c
}

func (c Challenge) WithAttemptsLeft(left int, now time.Time) Challenge {
	c.AttemptsLeft = left
	c.UpdatedAt = now
	return c
}

func (c Challenge) WithLockUntil(until *time.Time, now time.Time) Challenge {
	c.LockUntil = until
	c.UpdatedAt = now
	return c
}
