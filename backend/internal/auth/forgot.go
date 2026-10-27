package auth

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/MaryJane-09/noxorbit/backend/internal/email"
	"github.com/MaryJane-09/noxorbit/backend/internal/otp"
	"github.com/MaryJane-09/noxorbit/backend/internal/session"
	"github.com/MaryJane-09/noxorbit/backend/internal/user"
	"github.com/MaryJane-09/noxorbit/backend/internal/validate"
)

const resendCooldown = time.Minute 
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

	
	if existing, err := otpRepo.FindByEmail(addr, otp.PurposeReset); err == nil {
		issuedAt := existing.ExpiresAt.Add(-otp.ExpiryTime)
		if time.Since(issuedAt) < resendCooldown {
			return
		}
	}

	code, err := otp.Generate(6)
	if err != nil {
		log.Println(err)
		return
	}
	err = otpRepo.Create(otp.OTP{
		Email:     addr,
		Purpose:   otp.PurposeReset,
		Code:      code,
		ExpiresAt: time.Now().Add(otp.ExpiryTime),
	})
	if err != nil {
		log.Println(err)
		return
	}
	if err := sender.SendPasswordReset(u.Name, addr, code); err != nil {
		log.Println(err)
		_ = otpRepo.Delete(addr, otp.PurposeReset) 
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

		stored, err := otpRepo.FindByEmail(addr, otp.PurposeReset)
		if err != nil {
			writeError(w, http.StatusBadRequest, invalidCodeMessage)
			return
		}
		if time.Now().After(stored.ExpiresAt) {
			_ = otpRepo.Delete(addr, otp.PurposeReset)
			writeError(w, http.StatusBadRequest, invalidCodeMessage)
			return
		}
		
		if subtle.ConstantTimeCompare([]byte(code), []byte(stored.Code)) != 1 {
			writeError(w, http.StatusBadRequest, invalidCodeMessage)
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