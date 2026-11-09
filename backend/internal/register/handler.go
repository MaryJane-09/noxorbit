package register

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MaryJane-09/noxorbit/backend/internal/email"
	"github.com/MaryJane-09/noxorbit/backend/internal/otp"
	"github.com/MaryJane-09/noxorbit/backend/internal/user"
	"github.com/MaryJane-09/noxorbit/backend/internal/validate"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// RegisterHandler validates the submitted details and sends an email code.
// It does not persist the name or password before the email is verified.
func RegisterHandler(otpRepo *otp.OTPRepository, sender *email.EmailSender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
			return
		}
		defer r.Body.Close()

		var info user.User
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&info); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid registration request"})
			return
		}
		info.Name = strings.TrimSpace(info.Name)
		info.Email = strings.ToLower(strings.TrimSpace(info.Email))
		if err := validate.ValidateRegister(info); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			return
		}

		code, err := otp.Generate(6)
		if err != nil {
			log.Println("Generate registration code:", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Could not generate verification code"})
			return
		}
		err = otpRepo.Create(otp.OTP{
			Email:     info.Email,
			Purpose:   otp.PurposeRegister,
			Code:      code,
			ExpiresAt: time.Now().Add(otp.ExpiryTime),
			Verified:  false,
		})
		if err != nil {
			log.Println("Save registration code:", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Could not create verification code"})
			return
		}

		if err := sender.SendVerification(info.Name, info.Email, code); err != nil {
			log.Println("Send registration code:", err)
			_ = otpRepo.Delete(info.Email, otp.PurposeRegister)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Could not send verification email"})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Verification code generated successfully",
		})
	}
}
