package otp

import "time"

const (
	Time          = 5
	ExpiryTime    = Time * time.Minute
	IssueCooldown = time.Minute
	MaxAttempts   = 5
	CodeLength    = 6
)

const (
	PurposeRegister = "register"
	PurposeReset    = "reset"
)

type OTP struct {
	Email     string
	Purpose   string
	Code      string
	ExpiresAt time.Time
	Verified  bool
}
