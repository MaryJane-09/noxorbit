package membership

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/MaryJane-09/noxorbit/backend/internal/auth"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("workspace not found")

type Member struct {
	UserID uuid.UUID `json:"user_id"`
	Name string `json:"name"`
	Role string `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type Store interface {
	List(context.Context, uuid.UUID, uuid.UUID) ([]Member, error)
}

type Repository struct { pool *pgxpool.Pool }
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool:pool} }

// No global user directory: only users who share the requested workspace
// with the authenticated requester appear. Do not return emails or passwords.
func (r *Repository) List(ctx context.Context, actor, workspaceID uuid.UUID) ([]Member, error) {
	var visible bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM public.workspace_memberships
		  WHERE workspace_id=$1 AND user_id=$2)
	`,workspaceID,actor).Scan(&visible)
	if err != nil { return nil, err }
	if !visible { return nil, ErrNotFound }
	rows, err := r.pool.Query(ctx, `
		SELECT m.user_id, u.name, m.role, m.joined_at
		FROM public.workspace_memberships m
		JOIN public.users u ON u.id=m.user_id
		WHERE m.workspace_id=$1
		  AND EXISTS (SELECT 1 FROM public.workspace_memberships viewer
		    WHERE viewer.workspace_id=m.workspace_id AND viewer.user_id=$2)
		ORDER BY CASE WHEN m.role='owner' THEN 0 ELSE 1 END, m.joined_at, m.user_id
	`, workspaceID, actor)
	if err != nil { return nil, err }
	defer rows.Close()
	members := make([]Member,0)
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.UserID,&m.Name,&m.Role,&m.JoinedAt); err != nil { return nil,err }
		members=append(members,m)
	}
	return members,rows.Err()
}

type Handler struct { store Store }
func NewHandler(store Store) *Handler { return &Handler{store:store} }

func (h *Handler) List(w http.ResponseWriter,r *http.Request) {
	w.Header().Set("Content-Type","application/json")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow","GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error":"Method not allowed"})
		return
	}
	userID,ok:=auth.UserIDFromContext(r.Context())
	if !ok { w.WriteHeader(http.StatusUnauthorized); _=json.NewEncoder(w).Encode(map[string]string{"error":"Not logged in"});return }
	workspaceID,err:=uuid.FromString(r.PathValue("workspace_id"))
	if err != nil { w.WriteHeader(http.StatusBadRequest); _=json.NewEncoder(w).Encode(map[string]string{"error":"Invalid workspace ID"});return }
	members,err:=h.store.List(r.Context(),userID,workspaceID)
	if errors.Is(err,ErrNotFound) { w.WriteHeader(http.StatusNotFound); _=json.NewEncoder(w).Encode(map[string]string{"error":"Workspace not found"});return }
	if errors.Is(err,pgx.ErrNoRows) { w.WriteHeader(http.StatusNotFound); _=json.NewEncoder(w).Encode(map[string]string{"error":"Workspace not found"});return }
	if err != nil { log.Println("List workspace members:",err); w.WriteHeader(http.StatusInternalServerError); _=json.NewEncoder(w).Encode(map[string]string{"error":"Could not load members"});return }
	_ = json.NewEncoder(w).Encode(members)
}
