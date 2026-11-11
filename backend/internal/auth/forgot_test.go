package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MaryJane-09/noxorbit/backend/internal/otp"
)

type stubResetter func(context.Context, string, string, string) error

func (f stubResetter) ResetPassword(ctx context.Context, email, code, password string) error {
	return f(ctx, email, code, password)
}

func TestResetPasswordDoesNotRevealOTPState(t *testing.T) {
	cases := []error{
		otp.ErrNotFound, otp.ErrExpired, otp.ErrAlreadyUsed,
		otp.ErrInvalidCode, otp.ErrTooManyAttempts,
	}
	for _, cause := range cases {
		t.Run(cause.Error(), func(t *testing.T) {
			h := ResetPasswordHandler(stubResetter(func(context.Context, string, string, string) error {
				return cause
			}))
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(
				`{"email":"person@example.com","code":"abc234","newPassword":"VeryStrongPass123"}`))
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest || strings.TrimSpace(rec.Body.String()) != `{"error":"Invalid or expired code"}` {
				t.Fatalf("status %d and body %q must be generic", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestResetPasswordValidCodeCallsSingleTransactionStore(t *testing.T) {
	calls := 0
	h := ResetPasswordHandler(stubResetter(func(_ context.Context, email, code, password string) error {
		calls++
		if email != "person@example.com" || code != "abc234" || password != "VeryStrongPass123" {
			t.Fatalf("invalid reset parameters: %s %s", email, code)
		}
		return nil
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(
		`{"email":"Person@Example.Com","code":"abc234","newPassword":"VeryStrongPass123"}`))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || calls != 1 {
		t.Fatalf("reset status %d calls %d body %s", rec.Code, calls, rec.Body.String())
	}
}

func TestResetPasswordErrorsNeverLeakDatabaseDetails(t *testing.T) {
	h := ResetPasswordHandler(stubResetter(func(context.Context, string, string, string) error {
		return errors.New("database password must not leak")
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(
		`{"email":"person@example.com","code":"abc234","newPassword":"VeryStrongPass123"}`))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "database password") {
		t.Fatalf("response revealed error: %d %s", rec.Code, rec.Body.String())
	}
}
