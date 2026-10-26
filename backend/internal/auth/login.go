package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
	"github.com/MaryJane-09/noxorbit/backend/internal/register"
	"github.com/MaryJane-09/noxorbit/backend/internal/session"
	"github.com/MaryJane-09/noxorbit/backend/internal/user"
	"golang.org/x/crypto/bcrypt"
)

const cookieName = "session"
const sessionLifetime = 7 * 24 * time.Hour

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("not-a-real-password"), bcrypt.DefaultCost)

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, register.ErrorResponse{Error: message})
}

func LoginHandler(users *user.UserRepository, sessions *session.SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		defer r.Body.Close()

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request")
			return
		}
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))

		u, err := users.FindByEmail(req.Email)
		if err != nil {
			if errors.Is(err, user.ErrNotFound) {
				_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
				writeError(w, http.StatusUnauthorized, "Invalid email or password")
				return
			}
			log.Println(err)
			writeError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
			writeError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}

		token, err := session.NewToken()
		if err != nil {
			log.Println(err)
			writeError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		expiresAt := time.Now().Add(sessionLifetime)
		if err := sessions.Create(u.Id, session.HashToken(token), expiresAt); err != nil {
			log.Println(err)
			writeError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    token,
			Path:     "/",
			Expires:  expiresAt,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   false, 
		})
		writeJSON(w, http.StatusOK, map[string]string{"message": "Login successful"})
	}
}