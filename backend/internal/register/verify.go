package register

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/MaryJane-09/noxorbit/backend/internal/otp"
	"github.com/MaryJane-09/noxorbit/backend/internal/pending"
	"github.com/MaryJane-09/noxorbit/backend/internal/user"
)

type OTPVerification struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func VerifyOTPHandler(repo *user.UserRepository, otpRepo *otp.OTPRepository, pendingRepo *pending.PendingRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
			return
		}
		defer r.Body.Close()

		var info OTPVerification
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&info)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to decode request body"})
			return
		}
		storedOTP, err := otpRepo.FindByEmail(info.Email)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "No OTP was found for this email."})
			return
		}
		if time.Now().After(storedOTP.ExpiresAt) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "This OTP has expired"})
			err = otpRepo.Delete(info.Email)
			if err != nil {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete otp"})
				return
			}
			return
		}
		if info.Code != storedOTP.Code {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Incorrect OTP. No match found"})
			return
		}
		pendingUser, err := pendingRepo.FindByEmail(info.Email)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "User cannot be found"})
			return
		}
		err = repo.Create(pendingUser)
		if err != nil{
			if errors.Is(err, user.ErrEmailExists){
				w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResponse{Error: user.ErrEmailExists.Error()})
			return
			}
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create user"})
			return
		}

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

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "User created successfully",
		})
	}
}
