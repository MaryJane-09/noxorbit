package otp

import (
	"errors"
	"testing"
	"time"
)

func TestInspectCode(t *testing.T) {
	now := time.Date(2026, time.November, 9, 12, 0, 0, 0, time.UTC)
	unexpired := now.Add(5 * time.Minute)
	tests := []struct {
		name         string
		sent         string
		stored       string
		expires      time.Time
		verified     bool
		attempts     int
		wrong        bool
		expectedError error
	}{
		{
			name: "correct unexpired code accepted",
			sent: "abc234", stored: "abc234", expires: unexpired,
		},
		{
			name: "first incorrect guess increments",
			sent: "abc235", stored: "abc234", expires: unexpired,
			wrong: true, expectedError: ErrInvalidCode,
		},
		{
			name: "fifth wrong guess locks the code",
			sent: "abc235", stored: "abc234", expires: unexpired,
			attempts: 4, wrong: true, expectedError: ErrTooManyAttempts,
		},
		{
			name: "correct guess cannot bypass an existing lock",
			sent: "abc234", stored: "abc234", expires: unexpired,
			attempts: 5, expectedError: ErrTooManyAttempts,
		},
		{
			name: "consumed code cannot be reused",
			sent: "abc234", stored: "abc234", expires: unexpired,
			verified: true, expectedError: ErrTooManyAttempts,
		},
		{
			name: "expired code rejected",
			sent: "abc234", stored: "abc234", expires: now,
			expectedError: ErrExpired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrong, err := inspectCode(tt.sent, tt.stored, tt.expires, tt.verified, tt.attempts, now)
			if wrong != tt.wrong {
				t.Fatalf("wrong = %v, want %v", wrong, tt.wrong)
			}
			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("err = %v, want %v", err, tt.expectedError)
			}
		})
	}
}
