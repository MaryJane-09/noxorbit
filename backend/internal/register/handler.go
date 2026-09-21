package register

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/MaryJane-09/noxorbit/backend/internal/email"
	"github.com/MaryJane-09/noxorbit/backend/internal/otp"
	"github.com/MaryJane-09/noxorbit/backend/internal/pending"
	"github.com/MaryJane-09/noxorbit/backend/internal/user"
	"github.com/MaryJane-09/noxorbit/backend/internal/validate"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func RegisterHandler(repo *user.UserRepository, otpRepo *otp.OTPRepository, pendingRepo *pending.PendingRepository, sender *email.EmailSender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
			return
		}
		defer r.Body.Close()

		var info user.User
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&info)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to decode request body"})
			return
		}

		err = validate.ValidateRegister(info)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			return
		}

		err = pendingRepo.Create(info)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Pending user not saved"})
			return
		}
		code, err := otp.Generate(8)
		if err != nil {
			err = pendingRepo.Delete(info.Email)
			if err != nil {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete pending user"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Could not generate OTP"})
			return
		}

		NewOTP := otp.OTP{
			Email:     info.Email,
			Code:      code,
			ExpiresAt: time.Now().Add(otp.ExpiryTime),
			Verified:  false,
		}

		err = otpRepo.Create(NewOTP)
		if err != nil {
			err = pendingRepo.Delete(info.Email)
			if err != nil {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete pending user"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Could not create new OTP"})
			return
		}

		err = sender.SendVerification(info.Name, info.Email, code)
		if err != nil {
			err = otpRepo.Delete(info.Email)
			if err != nil {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete otp"})
				return
			}
			err = pendingRepo.Delete(info.Email)
			if err != nil {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete pending user"})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Verification code generated successfully",
		})
	}
}
