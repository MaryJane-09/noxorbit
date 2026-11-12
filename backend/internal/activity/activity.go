package activity

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

type Event struct {
    ID uuid.UUID `json:"id"`
    WorkspaceID uuid.UUID `json:"workspace_id"`
    WorkspaceName string `json:"workspace_name"`
    ActorID *uuid.UUID `json:"actor_id"`
    ActorName string `json:"actor_name"`
    Action string `json:"action"`
    EntityType string `json:"entity_type"`
    EntityID uuid.UUID `json:"entity_id"`
    EntityName string `json:"entity_name"`
    Details json.RawMessage `json:"details"`
    CreatedAt time.Time `json:"created_at"`
}

type Store interface {
    List(context.Context, uuid.UUID, uuid.UUID) ([]Event, error)
}

type Repository struct{ pool *pgxpool.Pool }
func NewRepository(pool *pgxpool.Pool)*Repository{return &Repository{pool:pool}}

// Explicit workspace membership check prevents nonmembers distinguishing a
// quiet workspace from one they cannot access. The query rechecks membership.
func (r *Repository) List(ctx context.Context, userID, workspaceID uuid.UUID)([]Event,error){
    var member bool
    err:=r.pool.QueryRow(ctx,`
        SELECT EXISTS (
            SELECT 1 FROM public.workspace_memberships
            WHERE workspace_id=$1 AND user_id=$2
        )
    `,workspaceID,userID).Scan(&member)
    if err!=nil{return nil,err}
    if !member{return nil,ErrNotFound}
    rows,err:=r.pool.Query(ctx,`
        SELECT a.id,a.workspace_id,w.name,a.actor_id,
               COALESCE(NULLIF(u.name,''),'Former member'),
               a.action,a.entity_type,a.entity_id,a.entity_name,
               a.details,a.created_at
        FROM public.activity_events a
        JOIN public.workspaces w ON w.id=a.workspace_id
        LEFT JOIN public.users u ON u.id=a.actor_id
        WHERE a.workspace_id=$1
          AND EXISTS (
              SELECT 1 FROM public.workspace_memberships m
              WHERE m.workspace_id=a.workspace_id AND m.user_id=$2
          )
        ORDER BY a.created_at DESC,a.id DESC
        LIMIT 100
    `,workspaceID,userID)
    if err!=nil{return nil,err}
    defer rows.Close()
    events:=make([]Event,0)
    for rows.Next(){
        var event Event
        if err:=rows.Scan(&event.ID,&event.WorkspaceID,&event.WorkspaceName,
            &event.ActorID,&event.ActorName,&event.Action,&event.EntityType,
            &event.EntityID,&event.EntityName,&event.Details,&event.CreatedAt);err!=nil{return nil,err}
        events=append(events,event)
    }
    return events,rows.Err()
}

type Handler struct{ store Store }
func NewHandler(store Store)*Handler{return &Handler{store:store}}

func (h *Handler) List(w http.ResponseWriter,r *http.Request) {
    w.Header().Set("Content-Type","application/json")
    if r.Method!=http.MethodGet{
        w.Header().Set("Allow","GET")
        w.WriteHeader(http.StatusMethodNotAllowed)
        _=json.NewEncoder(w).Encode(map[string]string{"error":"Method not allowed"})
        return
    }
    userID,ok:=auth.UserIDFromContext(r.Context())
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        _=json.NewEncoder(w).Encode(map[string]string{"error":"Not logged in"})
        return
    }
    workspaceID,err:=uuid.FromString(r.PathValue("workspace_id"))
    if err!=nil{
        w.WriteHeader(http.StatusBadRequest)
        _=json.NewEncoder(w).Encode(map[string]string{"error":"Invalid workspace ID"})
        return
    }
    items,err:=h.store.List(r.Context(),userID,workspaceID)
    if err!=nil {
        if errors.Is(err,ErrNotFound) {
            w.WriteHeader(http.StatusNotFound)
            _=json.NewEncoder(w).Encode(map[string]string{"error":"Workspace not found"})
            return
        }
        log.Println("List activity:",err)
        w.WriteHeader(http.StatusInternalServerError)
        _=json.NewEncoder(w).Encode(map[string]string{"error":"Could not load activity"})
        return
    }
    _=json.NewEncoder(w).Encode(items)
}
var _ Store = (*Repository)(nil)
