package auth

import (
	"errors"
	"log"
	"net/http"

	"github.com/MaryJane-09/noxorbit/backend/internal/user"
)

func MeHandler(users *user.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		id, ok := UserIDFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "Not logged in")
			return
		}

		u, err := users.FindByID(id)
		if err != nil {
			if errors.Is(err, user.ErrNotFound) {
				writeError(w, http.StatusUnauthorized, "Not logged in")
				return
			}
			log.Println(err)
			writeError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}

		writeJSON(w, http.StatusOK, user.UserResponse{Name: u.Name, Email: u.Email})
	}
}