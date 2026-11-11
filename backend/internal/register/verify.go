package register

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/MaryJane-09/noxorbit/backend/internal/otp"
	"github.com/MaryJane-09/noxorbit/backend/internal/user"
	"github.com/MaryJane-09/noxorbit/backend/internal/validate"
)

// Registration details arrive again with the emailed code, rather than being
// loaded from a mutable pending-registration row in the database.
type OTPVerification struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

func VerifyOTPHandler(repo *user.UserRepository, otpRepo *otp.OTPRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
			return
		}
		defer r.Body.Close()

		var req OTPVerification
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid verification request"})
			return
		}

		info := user.User{
			Name:     strings.TrimSpace(req.Name),
			Email:    strings.ToLower(strings.TrimSpace(req.Email)),
			Password: req.Password,
		}
		if err := validate.ValidateRegister(info); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			return
		}

		code := strings.TrimSpace(req.Code)
		if len(code) != otp.CodeLength {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Incorrect OTP. No match found"})
			return
		}

		// Verify holds a database row lock, counts wrong guesses, and marks
		// a successful code as used before it can be submitted again.
		if err := otpRepo.Verify(info.Email, otp.PurposeRegister, code); err != nil {
			switch {
			case errors.Is(err, otp.ErrNotFound):
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "No OTP was found for this email."})
			case errors.Is(err, otp.ErrExpired):
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "This OTP has expired"})
			case errors.Is(err, otp.ErrTooManyAttempts):
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Too many incorrect codes. Request a new one.", Code: "otp_locked"})
			case errors.Is(err, otp.ErrAlreadyUsed):
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "This code has already been used.", Code: "otp_used"})
			case errors.Is(err, otp.ErrInvalidCode):
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Incorrect OTP. No match found"})
			default:
				log.Println("Verify registration code:", err)
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Could not verify code"})
			}
			return
		}

		// Only the verified request determines the account's name and password.
		// UserRepository.Create hashes the password with bcrypt before saving it.
		if err := repo.Create(info); err != nil {
			if errors.Is(err, user.ErrEmailExists) {
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(ErrorResponse{Error: user.ErrEmailExists.Error()})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create user"})
			return
		}
		_ = otpRepo.Delete(info.Email, otp.PurposeRegister)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "User created successfully",
		})
	}
}
