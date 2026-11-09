package auth

import (
		"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"
	
	"github.com/MaryJane-09/noxorbit/backend/internal/email"
	"github.com/MaryJane-09/noxorbit/backend/internal/otp"
	"github.com/MaryJane-09/noxorbit/backend/internal/session"
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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request")
			return
		}
		addr := strings.ToLower(strings.TrimSpace(req.Email))
		if _, err := mail.ParseAddress(addr); err != nil {
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

	

	code, err := otp.Generate(6)
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

func ResetPasswordHandler(users *user.UserRepository, otpRepo *otp.OTPRepository, sessions *session.SessionRepository) http.HandlerFunc {
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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request")
			return
		}
		addr := strings.ToLower(strings.TrimSpace(req.Email))
		code := strings.TrimSpace(req.Code)

		if err := validate.ValidatePassword(req.NewPassword); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// OTP verification is atomic: five wrong guesses lock the code,
		// and a successful match consumes it before resetting the password.
		if err := otpRepo.Verify(addr, otp.PurposeReset, code); err != nil {
			switch {
			case errors.Is(err, otp.ErrTooManyAttempts):
				writeError(w, http.StatusTooManyRequests, "Too many incorrect codes. Request a new one.")
			case errors.Is(err, otp.ErrInvalidCode), errors.Is(err, otp.ErrNotFound), errors.Is(err, otp.ErrExpired):
				writeError(w, http.StatusBadRequest, invalidCodeMessage)
			default:
				log.Println("Verify reset code:", err)
				writeError(w, http.StatusInternalServerError, "Something went wrong")
			}
			return
		}

		u, err := users.FindByEmail(addr)
		if err != nil {
			if errors.Is(err, user.ErrNotFound) {
				writeError(w, http.StatusBadRequest, invalidCodeMessage)
				return
			}
			log.Println(err)
			writeError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		if err := users.UpdatePassword(u.Id, req.NewPassword); err != nil {
			log.Println(err)
			writeError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}
		if err := sessions.DeleteAllForUser(u.Id); err != nil {
			log.Println(err)
			writeError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}
		_ = otpRepo.Delete(addr, otp.PurposeReset)

		writeJSON(w, http.StatusOK, map[string]string{"message": "Password updated"})
	}
}