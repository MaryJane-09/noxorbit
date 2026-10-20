package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/MaryJane-09/noxorbit/backend/internal/register"
	"github.com/MaryJane-09/noxorbit/backend/internal/session"
	"github.com/MaryJane-09/noxorbit/backend/internal/user"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("not-a-real-password"), bcrypt.DefaultCost)

func LoginHandler(users *user.UserRepository, sessions *session.SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(register.ErrorResponse{Error: "Method not allowed"})
			return
		}
		defer r.Body.Close()

		var req LoginRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(register.ErrorResponse{Error: "Invalid Request"})
			return
		}
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		u, err := users.FindByEmail(req.Email)
		if err != nil {
			if errors.Is(err, user.ErrNotFound) {
				_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(register.ErrorResponse{Error: "Invalid email or password"})
				return
			}
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(register.ErrorResponse{Error: "Internal Server Error"})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password))
		if err != nil {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(register.ErrorResponse{Error: "Invalid email or password"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
	}
}
