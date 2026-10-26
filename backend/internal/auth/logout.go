package auth

import (
	"log"
	"net/http"

	"github.com/MaryJane-09/noxorbit/backend/internal/session"
)

func LogoutHandler(sessions *session.SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		if cookie, err := r.Cookie(cookieName); err == nil {
			if err := sessions.Delete(session.HashToken(cookie.Value)); err != nil {
				log.Println(err)
				writeError(w, http.StatusInternalServerError, "Something went wrong")
				return
			}
		}

		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    "",
			Path:     "/", 
			MaxAge:   -1, 
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		writeJSON(w, http.StatusOK, map[string]string{"message": "Logged out"})
	}
}