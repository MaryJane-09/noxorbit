package task

import (
    "context"
    "errors"
    "log"
    "net/http"
    "time"

    "github.com/MaryJane-09/noxorbit/backend/internal/auth"
    "github.com/gofrs/uuid"
    "github.com/jackc/pgx/v5"
)

// RecycledTask is a deleted task plus the project/workspace names needed by the bin.
type RecycledTask struct {
    Task
    ProjectName string `json:"project_name"`
    WorkspaceName string `json:"workspace_name"`
    DeletedAt time.Time `json:"deleted_at"`
    DeletedBy *uuid.UUID `json:"deleted_by"`
}

type RecycleStore interface {
    ListRecycled(context.Context, uuid.UUID) ([]RecycledTask, error)
    Restore(context.Context, uuid.UUID, uuid.UUID) error
    DeleteForever(context.Context, uuid.UUID, uuid.UUID) error
}

// Only tasks the member is eligible to restore are listed: owner or creator.
// Archived projects' tasks are displayed for context, but restoration requires
// the project to be active again.
func (r *Repository) ListRecycled(ctx context.Context, userID uuid.UUID) ([]RecycledTask, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT `+taskColumns+`,p.name,w.name,t.deleted_at,t.deleted_by
        FROM public.tasks AS t
        JOIN public.projects p ON p.id=t.project_id AND p.workspace_id=t.workspace_id
        JOIN public.workspaces w ON w.id=t.workspace_id
        JOIN public.workspace_memberships m ON m.workspace_id=t.workspace_id
            AND m.user_id=$1
        WHERE t.deleted_at IS NOT NULL AND (m.role='owner' OR t.created_by=$1)
        ORDER BY t.deleted_at DESC,t.id DESC
        LIMIT 200
    `,userID)
    if err!=nil{return nil,err}
    defer rows.Close()
    result:=make([]RecycledTask,0)
    for rows.Next() {
        var item RecycledTask
        if err:=rows.Scan(&item.ID,&item.ProjectID,&item.WorkspaceID,&item.Title,
            &item.Description,&item.Status,&item.Priority,&item.DueDate,
            &item.CreatedBy,&item.AssigneeID,&item.CreatedAt,&item.UpdatedAt,
            &item.ProjectName,&item.WorkspaceName,&item.DeletedAt,&item.DeletedBy);err!=nil{return nil,err}
        result=append(result,item)
    }
    return result,rows.Err()
}

// Restore is a single guarded UPDATE: membership, edit rights, project state,
// and deletion state are all verified within the same SQL statement.
func (r *Repository) Restore(ctx context.Context, userID, taskID uuid.UUID) error {
    var restored uuid.UUID
    err:=r.pool.QueryRow(ctx, `
        UPDATE public.tasks AS t
        SET deleted_at=NULL, deleted_by=NULL, activity_actor_id=$1, updated_at=NOW()
        FROM public.projects p
        WHERE t.id=$2 AND t.deleted_at IS NOT NULL
          AND p.id=t.project_id AND p.workspace_id=t.workspace_id
          AND p.status='active'
          AND EXISTS (
              SELECT 1 FROM public.workspace_memberships m
              WHERE m.user_id=$1 AND m.workspace_id=t.workspace_id
                AND (m.role='owner' OR t.created_by=$1)
          )
        RETURNING t.id
    `,userID,taskID).Scan(&restored)
    if errors.Is(err,pgx.ErrNoRows){return ErrNotFound}
    return err
}

// DeleteForever removes only an already-recycled task. Authorization is
// checked within the DELETE, so guessing task IDs never grants access.
// Unlike restoration, it also works when the parent project is archived.
func (r *Repository) DeleteForever(ctx context.Context, userID, taskID uuid.UUID) error {
    var deleted uuid.UUID
    err := r.pool.QueryRow(ctx, `
        WITH removed AS (
            DELETE FROM public.tasks AS t
            USING public.projects AS p
            WHERE t.id = $2 AND t.deleted_at IS NOT NULL
              AND p.id = t.project_id AND p.workspace_id = t.workspace_id
              AND EXISTS (
                  SELECT 1 FROM public.workspace_memberships AS m
                  WHERE m.workspace_id = t.workspace_id AND m.user_id = $1
                    AND (m.role = 'owner' OR t.created_by = $1)
              )
            RETURNING t.id, t.workspace_id, t.title
        ), logged AS (
            INSERT INTO public.activity_events
                (workspace_id, actor_id, action, entity_type, entity_id, entity_name)
            SELECT workspace_id, $1, 'task.deleted_forever', 'task', id, title
            FROM removed
            RETURNING entity_id
        )
        SELECT entity_id FROM logged
    `, userID, taskID).Scan(&deleted)
    if errors.Is(err, pgx.ErrNoRows) { return ErrNotFound }
    return err
}

type RecycleHandler struct{ store RecycleStore }
func NewRecycleHandler(store RecycleStore)*RecycleHandler{return &RecycleHandler{store:store}}

func (h *RecycleHandler) List(w http.ResponseWriter,r *http.Request) {
    if r.Method!=http.MethodGet{
        w.Header().Set("Allow","GET")
        fail(w,http.StatusMethodNotAllowed,"Method not allowed")
        return
    }
    userID,ok:=auth.UserIDFromContext(r.Context())
    if !ok{fail(w,http.StatusUnauthorized,"Not logged in");return}
    items,err:=h.store.ListRecycled(r.Context(),userID)
    if err!=nil{log.Println("Recycle bin list:",err);fail(w,http.StatusInternalServerError,"Could not load recycle bin");return}
    respond(w,http.StatusOK,items)
}

func (h *RecycleHandler) Restore(w http.ResponseWriter,r *http.Request){
    if r.Method!=http.MethodPost{
        w.Header().Set("Allow","POST")
        fail(w,http.StatusMethodNotAllowed,"Method not allowed")
        return
    }
    userID,ok:=auth.UserIDFromContext(r.Context())
    if !ok{fail(w,http.StatusUnauthorized,"Not logged in");return}
    taskID,err:=uuid.FromString(r.PathValue("task_id"))
    if err!=nil{fail(w,http.StatusBadRequest,"Invalid task ID");return}
    if err=h.store.Restore(r.Context(),userID,taskID);err!=nil{
        if errors.Is(err,ErrNotFound){fail(w,http.StatusNotFound,"Task cannot be restored. Check its project and your permissions.");return}
        log.Println("Recycle bin restore:",err)
        fail(w,http.StatusInternalServerError,"Could not restore task")
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v1/recycle-bin/tasks/{task_id} permanently deletes a task
// that has already been moved to the recycle bin.
func (h *RecycleHandler) DeleteForever(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        w.Header().Set("Allow", "DELETE")
        fail(w, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }
    userID, ok := auth.UserIDFromContext(r.Context())
    if !ok { fail(w, http.StatusUnauthorized, "Not logged in"); return }
    taskID, err := uuid.FromString(r.PathValue("task_id"))
    if err != nil { fail(w, http.StatusBadRequest, "Invalid task ID"); return }
    if err := h.store.DeleteForever(r.Context(), userID, taskID); err != nil {
        if errors.Is(err, ErrNotFound) {
            fail(w, http.StatusNotFound, "Task not found in your recycle bin")
            return
        }
        log.Println("Recycle bin permanent deletion:", err)
        fail(w, http.StatusInternalServerError, "Could not permanently delete task")
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

var _ RecycleStore = (*Repository)(nil)
