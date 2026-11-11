package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"

	"github.com/MaryJane-09/noxorbit/backend/internal/auth"
	"github.com/gofrs/uuid"
)

const maxCreateBodyBytes = 64 << 10

// Store keeps HTTP validation separate from the PostgreSQL implementation,
// allowing handlers to be tested without a live database.
type Store interface {
	Create(ctx context.Context, ownerID uuid.UUID, name, purpose string) (Workspace, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]Workspace, error)
	GetForUser(ctx context.Context, userID, workspaceID uuid.UUID) (Workspace, error)
}

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func respondJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// Collection implements POST/GET /api/v1/workspaces.
// The caller's user ID is taken only from the authenticated session context.
func (h *Handler) Collection(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	switch r.Method {
	case http.MethodGet:
		workspaces, err := h.store.ListForUser(r.Context(), userID)
		if err != nil {
			log.Println("List workspaces:", err)
			respondError(w, http.StatusInternalServerError, "Could not load workspaces")
			return
		}
		respondJSON(w, http.StatusOK, workspaces)

	case http.MethodPost:
		contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || contentType != "application/json" {
			respondError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
			return
		}

		defer r.Body.Close()
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxCreateBodyBytes))
		decoder.DisallowUnknownFields()

		var input CreateRequest
		if err := decoder.Decode(&input); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid workspace request")
			return
		}
		// A second JSON object (or any trailing garbage) is not allowed.
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			respondError(w, http.StatusBadRequest, "Invalid workspace request")
			return
		}

		input, err = NormalizeCreate(input)
		if err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		ws, err := h.store.Create(r.Context(), userID, input.Name, input.Purpose)
		if err != nil {
			log.Println("Create workspace:", err)
			respondError(w, http.StatusInternalServerError, "Could not create workspace")
			return
		}
		respondJSON(w, http.StatusCreated, ws)

	default:
		w.Header().Set("Allow", "GET, POST")
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Detail implements GET /api/v1/workspaces/{id}.
// A valid UUID is not enough; membership is enforced by GetForUser's JOIN.
func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Not logged in")
		return
	}
	workspaceID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid workspace ID")
		return
	}

	ws, err := h.store.GetForUser(r.Context(), userID, workspaceID)
	if errors.Is(err, ErrNotFound) {
		respondError(w, http.StatusNotFound, "Workspace not found")
		return
	}
	if err != nil {
		log.Println("Get workspace:", err)
		respondError(w, http.StatusInternalServerError, "Could not load workspace")
		return
	}
	respondJSON(w, http.StatusOK, ws)
}
