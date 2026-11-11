package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MaryJane-09/noxorbit/backend/internal/email"
	"github.com/MaryJane-09/noxorbit/backend/internal/otp"
	"github.com/MaryJane-09/noxorbit/backend/internal/user"
	"github.com/MaryJane-09/noxorbit/backend/internal/validate"
)

const invalidCodeMessage = "Invalid or expired code"

func ForgotPasswordHandler(users *user.UserRepository, otpRepo *otp.OTPRepository, sender *email.EmailSender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		defer r.Body.Close()

		var req struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request")
			return
		}
		addr := strings.ToLower(strings.TrimSpace(req.Email))
		if err := validate.ValidateEmail(addr); err != nil {
			writeError(w, http.StatusBadRequest, "Enter a valid email address")
			return
		}

		go sendResetCode(users, otpRepo, sender, addr)

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "If an account exists for that email, a code is on its way.",
		})
	}
}

func sendResetCode(users *user.UserRepository, otpRepo *otp.OTPRepository, sender *email.EmailSender, addr string) {
	u, err := users.FindByEmail(addr)
	if err != nil {
		if !errors.Is(err, user.ErrNotFound) {
			log.Println(err)
		}
		return
	}


	code, err := otp.Generate(otp.CodeLength)
	if err != nil {
		log.Println(err)
		return
	}
	err = otpRepo.Issue(otp.OTP{
		Email:     addr,
		Purpose:   otp.PurposeReset,
		Code:      code,
		ExpiresAt: time.Now().Add(otp.ExpiryTime),
	})
	if err != nil {
		if !errors.Is(err, otp.ErrCooldown) {
			log.Println("Save reset code:", err)
		}
		return
	}
	if err := sender.SendPasswordReset(u.Name, addr, code); err != nil {
		log.Println(err)
		_ = otpRepo.DeleteIfCode(addr, otp.PurposeReset, code)
	}
}

// passwordResetter keeps the HTTP policy testable without live PostgreSQL.
// The OTP repository carries out successful resets in one transaction.
type passwordResetter interface {
	ResetPassword(ctx context.Context, email, code, newPassword string) error
}

func ResetPasswordHandler(resetter passwordResetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		defer r.Body.Close()

		var req struct {
			Email       string `json:"email"`
			Code        string `json:"code"`
			NewPassword string `json:"newPassword"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		addr := strings.ToLower(strings.TrimSpace(req.Email))
		code := strings.TrimSpace(req.Code)
		if err := validate.ValidatePassword(req.NewPassword); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := validate.ValidateEmail(addr); err != nil || len(code) != otp.CodeLength {
			writeError(w, http.StatusBadRequest, invalidCodeMessage)
			return
		}

		err := resetter.ResetPassword(r.Context(), addr, code, req.NewPassword)
		switch {
		case err == nil:
			writeJSON(w, http.StatusOK, map[string]string{"message": "Password updated"})
		case errors.Is(err, otp.ErrTooManyAttempts),
			errors.Is(err, otp.ErrAlreadyUsed),
			errors.Is(err, otp.ErrInvalidCode),
			errors.Is(err, otp.ErrExpired),
			errors.Is(err, otp.ErrNotFound):
			// Uniform public response prevents different OTP states from
			// revealing whether a password-reset code/account exists.
			writeError(w, http.StatusBadRequest, invalidCodeMessage)
		default:
			log.Println("Reset password:", err)
			writeError(w, http.StatusInternalServerError, "Something went wrong")
		}
	}
}
