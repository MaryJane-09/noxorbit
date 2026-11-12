package task

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"time"

	"github.com/MaryJane-09/noxorbit/backend/internal/auth"
	"github.com/gofrs/uuid"
)

const maxTaskBodyBytes = 16 << 10

type Store interface {
	Create(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, DetailsInput) (Task, error)
	List(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]Task, error)
	Get(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) (Task, error)
	UpdateDetails(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, DetailsInput) (Task, error)
	UpdateStatus(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, string) (Task, error)
	Delete(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) error
}

type Handler struct { store Store }
func NewHandler(store Store) *Handler { return &Handler{store: store} }

func respond(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
func fail(w http.ResponseWriter, code int, message string) {
	respond(w, code, map[string]string{"error": message})
}
func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		fail(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return false
	}
	defer r.Body.Close()
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxTaskBodyBytes))
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		fail(w, http.StatusBadRequest, "Invalid task request")
		return false
	}
	var extra any
	if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
		fail(w, http.StatusBadRequest, "Invalid task request")
		return false
	}
	return true
}

func routeIDs(w http.ResponseWriter, r *http.Request, withTask bool) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		fail(w, http.StatusUnauthorized, "Not logged in")
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	wid, err := uuid.FromString(r.PathValue("workspace_id"))
	if err != nil {
		fail(w, http.StatusBadRequest, "Invalid workspace ID")
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	pid, err := uuid.FromString(r.PathValue("project_id"))
	if err != nil {
		fail(w, http.StatusBadRequest, "Invalid project ID")
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	if !withTask { return userID, wid, pid, uuid.Nil, true }
	tid, err := uuid.FromString(r.PathValue("task_id"))
	if err != nil {
		fail(w, http.StatusBadRequest, "Invalid task ID")
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	return userID, wid, pid, tid, true
}

func repositoryError(w http.ResponseWriter, err error) bool {
	if err == nil { return false }
	switch {
	case errors.Is(err, ErrNotFound):
		fail(w, http.StatusNotFound, "Task or project not found")
	case errors.Is(err, ErrForbidden):
		fail(w, http.StatusForbidden, "You cannot change this task")
	case errors.Is(err, ErrArchived):
		fail(w, http.StatusConflict, "Archived projects cannot be changed")
	case errors.Is(err, ErrInvalidAssignee):
		fail(w, http.StatusBadRequest, ErrInvalidAssignee.Error())
	default:
		log.Println("Task repository:", err)
		fail(w, http.StatusInternalServerError, "Could not process task request")
	}
	return true
}

// GET and POST /api/v1/workspaces/{workspace_id}/projects/{project_id}/tasks
func (h *Handler) Collection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		fail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	uid, wid, pid, _, ok := routeIDs(w, r, false)
	if !ok { return }
	if r.Method == http.MethodGet {
		items, err := h.store.List(r.Context(), uid, wid, pid)
		if repositoryError(w, err) { return }
		respond(w, http.StatusOK, items)
		return
	}
	var input DetailsInput
	if !decode(w, r, &input) { return }
	input, err := NormalizeDetails(input)
	if err != nil { fail(w, http.StatusBadRequest, err.Error()); return }
	if err := ValidateNewDueDate(input.DueDate, nil, time.Now()); err != nil { fail(w, http.StatusBadRequest, err.Error()); return }
	item, err := h.store.Create(r.Context(), uid, wid, pid, input)
	if repositoryError(w, err) { return }
	respond(w, http.StatusCreated, item)
}

// GET, PATCH and DELETE /api/v1/workspaces/{workspace_id}/projects/{project_id}/tasks/{task_id}
func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPatch && r.Method != http.MethodDelete {
		w.Header().Set("Allow", "GET, PATCH, DELETE")
		fail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	uid, wid, pid, tid, ok := routeIDs(w, r, true)
	if !ok { return }
	switch r.Method {
	case http.MethodGet:
		item, err := h.store.Get(r.Context(), uid, wid, pid, tid)
		if repositoryError(w, err) { return }
		respond(w, http.StatusOK, item)
	case http.MethodDelete:
		if repositoryError(w, h.store.Delete(r.Context(), uid, wid, pid, tid)) { return }
		w.WriteHeader(http.StatusNoContent)
	case http.MethodPatch:
		var input DetailsInput
		if !decode(w, r, &input) { return }
		input, err := NormalizeDetails(input)
		if err != nil { fail(w, http.StatusBadRequest, err.Error()); return }
		previous, err := h.store.Get(r.Context(), uid, wid, pid, tid)
		if repositoryError(w, err) { return }
		if err := ValidateNewDueDate(input.DueDate, previous.DueDate, time.Now()); err != nil { fail(w, http.StatusBadRequest, err.Error()); return }
		item, err := h.store.UpdateDetails(r.Context(), uid, wid, pid, tid, input)
		if repositoryError(w, err) { return }
		respond(w, http.StatusOK, item)
	}
}

// PATCH /api/v1/workspaces/{workspace_id}/projects/{project_id}/tasks/{task_id}/status
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", "PATCH")
		fail(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	uid, wid, pid, tid, ok := routeIDs(w, r, true)
	if !ok { return }
	var input StatusInput
	if !decode(w, r, &input) { return }
	status, err := NormalizeStatus(input.Status)
	if err != nil { fail(w, http.StatusBadRequest, err.Error()); return }
	item, err := h.store.UpdateStatus(r.Context(), uid, wid, pid, tid, status)
	if repositoryError(w, err) { return }
	respond(w, http.StatusOK, item)
}
