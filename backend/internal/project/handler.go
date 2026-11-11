package project

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

const maxProjectBodyBytes = 16 << 10

type Store interface {
	Create(ctx context.Context, userID, workspaceID uuid.UUID, input Input) (Project, error)
	ListForUser(ctx context.Context, userID, workspaceID uuid.UUID) ([]Project, error)
	GetForUser(ctx context.Context, userID, workspaceID, projectID uuid.UUID) (Project, error)
	UpdateForMember(ctx context.Context, userID, workspaceID, projectID uuid.UUID, input Input) (Project, error)
	UpdateStatusForOwner(ctx context.Context, userID, workspaceID, projectID uuid.UUID, status string) (Project, error)
}

type Handler struct { store Store }

func NewHandler(store Store) *Handler { return &Handler{store: store} }

func respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
func fail(w http.ResponseWriter, status int, message string) {
	respond(w, status, map[string]string{"error": message})
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		fail(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return false
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxProjectBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		fail(w, http.StatusBadRequest, "Invalid project request")
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		fail(w, http.StatusBadRequest, "Invalid project request")
		return false
	}
	return true
}

func requestContext(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		fail(w, http.StatusUnauthorized, "Not logged in")
		return uuid.Nil, uuid.Nil, false
	}
	workspaceID, err := uuid.FromString(r.PathValue("workspace_id"))
	if err != nil {
		fail(w, http.StatusBadRequest, "Invalid workspace ID")
		return uuid.Nil, uuid.Nil, false
	}
	return userID, workspaceID, true
}

func projectID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.FromString(r.PathValue("project_id"))
	if err != nil {
		fail(w, http.StatusBadRequest, "Invalid project ID")
		return uuid.Nil, false
	}
	return id, true
}

func repositoryError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, ErrWorkspaceNotFound):
		fail(w, http.StatusNotFound, "Workspace not found")
	case errors.Is(err, ErrNotFound):
		fail(w, http.StatusNotFound, "Project not found")
	case errors.Is(err, ErrForbidden):
		fail(w, http.StatusForbidden, "Only the workspace owner can archive or restore a project")
	case err != nil:
		log.Println("Project repository:", err)
		fail(w, http.StatusInternalServerError, "Could not process project request")
	default:
		return false
	}
	return true
}

// GET and POST /api/v1/workspaces/{workspace_id}/projects
func (h *Handler) Collection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		fail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	userID, wid, ok := requestContext(w, r)
	if !ok { return }

	if r.Method == http.MethodGet {
		projects, err := h.store.ListForUser(r.Context(), userID, wid)
		if repositoryError(w, err) { return }
		respond(w, http.StatusOK, projects)
		return
	}

	var input Input
	if !decode(w, r, &input) { return }
	input, err := Normalize(input)
	if err != nil { fail(w, http.StatusBadRequest, err.Error()); return }
	p, err := h.store.Create(r.Context(), userID, wid, input)
	if repositoryError(w, err) { return }
	respond(w, http.StatusCreated, p)
}

// GET and PATCH /api/v1/workspaces/{workspace_id}/projects/{project_id}
func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPatch {
		w.Header().Set("Allow", "GET, PATCH")
		fail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	userID, wid, ok := requestContext(w, r)
	if !ok { return }
	pid, ok := projectID(w, r)
	if !ok { return }

	if r.Method == http.MethodGet {
		p, err := h.store.GetForUser(r.Context(), userID, wid, pid)
		if repositoryError(w, err) { return }
		respond(w, http.StatusOK, p)
		return
	}

	var input Input
	if !decode(w, r, &input) { return }
	input, err := Normalize(input)
	if err != nil { fail(w, http.StatusBadRequest, err.Error()); return }
	p, err := h.store.UpdateForMember(r.Context(), userID, wid, pid, input)
	if repositoryError(w, err) { return }
	respond(w, http.StatusOK, p)
}

// PATCH /api/v1/workspaces/{workspace_id}/projects/{project_id}/status
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", "PATCH")
		fail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	userID, wid, ok := requestContext(w, r)
	if !ok { return }
	pid, ok := projectID(w, r)
	if !ok { return }
	var input StatusInput
	if !decode(w, r, &input) { return }
	status, err := NormalizeStatus(input.Status)
	if err != nil { fail(w, http.StatusBadRequest, err.Error()); return }
	p, err := h.store.UpdateStatusForOwner(r.Context(), userID, wid, pid, status)
	if repositoryError(w, err) { return }
	respond(w, http.StatusOK, p)
}
