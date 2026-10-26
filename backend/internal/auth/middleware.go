package auth

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/MaryJane-09/noxorbit/backend/internal/session"
	"github.com/gofrs/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

func RequireAuth(sessions *session.SessionRepository, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Not logged in")
			return
		}

		userID, err := sessions.FindUserID(session.HashToken(cookie.Value))
		if err != nil {
			if errors.Is(err, session.ErrNotFound) {
				writeError(w, http.StatusUnauthorized, "Not logged in")
				return
			}
			log.Println(err)
			writeError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}