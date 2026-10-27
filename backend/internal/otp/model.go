package otp

import (
	"time"
)

const ExpiryTime = Time * time.Minute
const Time = 5

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
