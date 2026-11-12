package task

import (
    "context"
    "encoding/json"
    "log"
    "net/http"

    "github.com/MaryJane-09/noxorbit/backend/internal/auth"
    "github.com/gofrs/uuid"
)

// AssignedTask includes the project and workspace context required by My Tasks.
// It intentionally excludes other users' assignments.
type AssignedTask struct {
    Task
    ProjectName string `json:"project_name"`
    WorkspaceName string `json:"workspace_name"`
    ProjectStatus string `json:"project_status"`
}

type AssignedStore interface {
    ListAssigned(context.Context, uuid.UUID) ([]AssignedTask, error)
}

func (r *Repository) ListAssigned(ctx context.Context, userID uuid.UUID) ([]AssignedTask, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT `+taskColumns+`, p.name, w.name, p.status
        FROM public.tasks AS t
        JOIN public.projects AS p ON p.id=t.project_id AND p.workspace_id=t.workspace_id
        JOIN public.workspaces AS w ON w.id=t.workspace_id
        JOIN public.workspace_memberships AS m ON m.workspace_id=t.workspace_id
            AND m.user_id=$1
        WHERE t.assignee_id=$1
        ORDER BY (t.status='done'), t.due_date ASC NULLS LAST,
            t.updated_at DESC, t.id DESC
    `,userID)
    if err!=nil{return nil,err}
    defer rows.Close()
    tasks:=make([]AssignedTask,0)
    for rows.Next(){
        var item AssignedTask
        if err:=rows.Scan(&item.ID,&item.ProjectID,&item.WorkspaceID,&item.Title,
            &item.Description,&item.Status,&item.Priority,&item.DueDate,
            &item.CreatedBy,&item.AssigneeID,&item.CreatedAt,&item.UpdatedAt,
            &item.ProjectName,&item.WorkspaceName,&item.ProjectStatus);err!=nil{return nil,err}
        tasks=append(tasks,item)
    }
    return tasks,rows.Err()
}

type AssignedHandler struct { store AssignedStore }
func NewAssignedHandler(store AssignedStore)*AssignedHandler{return &AssignedHandler{store:store}}

// GET /api/v1/my-tasks; the caller identity always comes from the session.
func (h *AssignedHandler) List(w http.ResponseWriter,r *http.Request){
    w.Header().Set("Content-Type","application/json")
    if r.Method!=http.MethodGet {
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
    items,err:=h.store.ListAssigned(r.Context(),userID)
    if err!=nil{
        log.Println("List assigned tasks:",err)
        w.WriteHeader(http.StatusInternalServerError)
        _=json.NewEncoder(w).Encode(map[string]string{"error":"Could not load assigned tasks"})
        return
    }
    _=json.NewEncoder(w).Encode(items)
}

// Explicit compile-time check, separate from database integration coverage.
var _ AssignedStore = (*Repository)(nil)
