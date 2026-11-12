package task

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct { pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

const taskColumns = `t.id, t.project_id, t.workspace_id, t.title, t.description,
	t.status, t.priority, t.due_date::text, t.created_by, t.assignee_id,
	t.created_at, t.updated_at`

func scanTask(row pgx.Row) (Task, error) {
	var t Task
	err := row.Scan(&t.ID, &t.ProjectID, &t.WorkspaceID, &t.Title,
		&t.Description, &t.Status, &t.Priority, &t.DueDate, &t.CreatedBy,
		&t.AssigneeID, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (r *Repository) accessibleProject(ctx context.Context, userID, workspaceID, projectID uuid.UUID) (string, error) {
	var status string
	err := r.pool.QueryRow(ctx, `
		SELECT p.status FROM public.projects p
		JOIN public.workspace_memberships m
		  ON m.workspace_id = p.workspace_id AND m.user_id = $1
		WHERE p.workspace_id = $2 AND p.id = $3
	`, userID, workspaceID, projectID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) { return "", ErrNotFound }
	return status, err
}

func (r *Repository) classifyWriteFailure(ctx context.Context, userID, workspaceID, projectID, taskID uuid.UUID) error {
	p, err := r.Get(ctx, userID, workspaceID, projectID, taskID)
	if err != nil { return err }
	_ = p
	status, err := r.accessibleProject(ctx, userID, workspaceID, projectID)
	if err != nil { return err }
	if status != "active" { return ErrArchived }
	return ErrForbidden
}

func mapAssigneeFK(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "tasks_workspace_assignee_fk" {
		return ErrInvalidAssignee
	}
	return err
}

// Any workspace member can create tasks in active projects and assign them to
// current members. The assignee and membership checks run in the INSERT itself.
func (r *Repository) Create(ctx context.Context, userID, workspaceID, projectID uuid.UUID, in DetailsInput) (Task, error) {
	t, err := scanTask(r.pool.QueryRow(ctx, `
		INSERT INTO public.tasks AS t
			(project_id, workspace_id, title, description, priority, due_date, created_by, assignee_id)
		SELECT p.id, p.workspace_id, $4, $5, $6, $7::date, $1, $8
		FROM public.projects p
		WHERE p.id = $3 AND p.workspace_id = $2 AND p.status = 'active'
		  AND EXISTS (
			SELECT 1 FROM public.workspace_memberships m
			WHERE m.workspace_id = p.workspace_id AND m.user_id = $1
		  )
		  AND ($7::date IS NULL OR $7::date >= (NOW() AT TIME ZONE 'Africa/Lagos')::date)
		  AND ($8::uuid IS NULL OR EXISTS (
			SELECT 1 FROM public.workspace_memberships a
			WHERE a.workspace_id = p.workspace_id AND a.user_id = $8
		  ))
		RETURNING `+taskColumns+`
	`, userID, workspaceID, projectID, in.Title, in.Description, in.Priority, in.DueDate, in.AssigneeID))
	if err == nil { return t, nil }
	if !errors.Is(err, pgx.ErrNoRows) { return Task{}, mapAssigneeFK(err) }
	status, lookupErr := r.accessibleProject(ctx, userID, workspaceID, projectID)
	if lookupErr != nil { return Task{}, lookupErr }
	if status != "active" { return Task{}, ErrArchived }
	if err := ValidateNewDueDate(in.DueDate, nil, time.Now()); err != nil { return Task{}, err }
	return Task{}, ErrInvalidAssignee
}

// All workspace members can see tasks in their workspace's projects.
func (r *Repository) List(ctx context.Context, userID, workspaceID, projectID uuid.UUID) ([]Task, error) {
	if _, err := r.accessibleProject(ctx, userID, workspaceID, projectID); err != nil { return nil, err }
	rows, err := r.pool.Query(ctx, `
		SELECT `+taskColumns+` FROM public.tasks t
		WHERE t.workspace_id = $2 AND t.project_id = $3 AND t.deleted_at IS NULL
		  AND EXISTS (SELECT 1 FROM public.workspace_memberships m
		              WHERE m.workspace_id = t.workspace_id AND m.user_id = $1)
		ORDER BY t.created_at DESC, t.id DESC
	`, userID, workspaceID, projectID)
	if err != nil { return nil, err }
	defer rows.Close()
	items := make([]Task, 0)
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.WorkspaceID, &t.Title,
			&t.Description, &t.Status, &t.Priority, &t.DueDate,
			&t.CreatedBy, &t.AssigneeID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (r *Repository) Get(ctx context.Context, userID, workspaceID, projectID, taskID uuid.UUID) (Task, error) {
	t, err := scanTask(r.pool.QueryRow(ctx, `
		SELECT `+taskColumns+` FROM public.tasks t
		WHERE t.id = $4 AND t.project_id = $3 AND t.workspace_id = $2 AND t.deleted_at IS NULL
		  AND EXISTS (SELECT 1 FROM public.workspace_memberships m
		              WHERE m.workspace_id = t.workspace_id AND m.user_id = $1)
	`, userID, workspaceID, projectID, taskID))
	if errors.Is(err, pgx.ErrNoRows) { return Task{}, ErrNotFound }
	return t, err
}

// Only task creators and workspace owners can change task details, priorities,
// assignees, and deadlines. Assignments must target current workspace members.
func (r *Repository) UpdateDetails(ctx context.Context, userID, workspaceID, projectID, taskID uuid.UUID, in DetailsInput) (Task, error) {
	t, err := scanTask(r.pool.QueryRow(ctx, `
		UPDATE public.tasks AS t SET
			title = $5, description = $6, priority = $7,
			due_date = $8::date, assignee_id = $9, activity_actor_id = $1, updated_at = NOW()
		FROM public.projects p
		WHERE t.id = $4 AND t.project_id = $3 AND t.workspace_id = $2
		  AND t.deleted_at IS NULL
		  AND p.id = t.project_id AND p.workspace_id = t.workspace_id
		  AND p.status = 'active'
		  AND EXISTS (
			SELECT 1 FROM public.workspace_memberships m
			WHERE m.workspace_id = t.workspace_id AND m.user_id = $1
			  AND (m.role = 'owner' OR t.created_by = $1)
		  )
		  AND ($8::date IS NULL OR $8::date = t.due_date OR $8::date >= (NOW() AT TIME ZONE 'Africa/Lagos')::date)
		  AND ($9::uuid IS NULL OR EXISTS (
			SELECT 1 FROM public.workspace_memberships a
			WHERE a.workspace_id = t.workspace_id AND a.user_id = $9
		  ))
		RETURNING `+taskColumns+`
	`, userID, workspaceID, projectID, taskID, in.Title, in.Description, in.Priority, in.DueDate, in.AssigneeID))
	if err == nil { return t, nil }
	if !errors.Is(err, pgx.ErrNoRows) { return Task{}, mapAssigneeFK(err) }
	if failure := r.classifyWriteFailure(ctx, userID, workspaceID, projectID, taskID); !errors.Is(failure, ErrForbidden) {
		return Task{}, failure
	}
	// Report an invalid assignee only to authorized editors, without leaking
	// membership details to someone who cannot edit this task.
	var canEdit bool
	err = r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM public.tasks t
			JOIN public.workspace_memberships m ON m.workspace_id=t.workspace_id
			WHERE t.id=$4 AND t.workspace_id=$2 AND t.project_id=$3 AND t.deleted_at IS NULL
			  AND m.user_id=$1 AND (m.role='owner' OR t.created_by=$1)
		)
	`, userID, workspaceID, projectID, taskID).Scan(&canEdit)
	if err != nil { return Task{}, err }
	if canEdit {
        current, getErr := r.Get(ctx, userID, workspaceID, projectID, taskID)
        if getErr != nil { return Task{}, getErr }
        if err := ValidateNewDueDate(in.DueDate, current.DueDate, time.Now()); err != nil { return Task{}, err }
        return Task{}, ErrInvalidAssignee
    }
	return Task{}, ErrForbidden
}

// Only owner, creator, or current assignee can change a task's status.
func (r *Repository) UpdateStatus(ctx context.Context, userID, workspaceID, projectID, taskID uuid.UUID, status string) (Task, error) {
	t, err := scanTask(r.pool.QueryRow(ctx, `
		UPDATE public.tasks AS t SET status = $5, activity_actor_id = $1, updated_at = NOW()
		FROM public.projects p
		WHERE t.id = $4 AND t.project_id = $3 AND t.workspace_id = $2
		  AND t.deleted_at IS NULL
		  AND p.id = t.project_id AND p.workspace_id = t.workspace_id
		  AND p.status = 'active'
		  AND EXISTS (
			SELECT 1 FROM public.workspace_memberships m
			WHERE m.workspace_id = t.workspace_id AND m.user_id = $1
			  AND (m.role = 'owner' OR t.created_by = $1 OR t.assignee_id = $1)
		  )
		RETURNING `+taskColumns+`
	`, userID, workspaceID, projectID, taskID, status))
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, r.classifyWriteFailure(ctx, userID, workspaceID, projectID, taskID)
	}
	return t, err
}

// Delete moves a task into the recycle bin. Only the task creator or workspace
// owner can remove a task, and archived projects remain read-only.
func (r *Repository) Delete(ctx context.Context, userID, workspaceID, projectID, taskID uuid.UUID) error {
    var deleted uuid.UUID
    err := r.pool.QueryRow(ctx, `
        UPDATE public.tasks AS t
        SET deleted_at = NOW(), deleted_by = $1, updated_at = NOW()
        FROM public.projects p
        WHERE t.id = $4 AND t.project_id = $3 AND t.workspace_id = $2
          AND t.deleted_at IS NULL
          AND p.id = t.project_id AND p.workspace_id = t.workspace_id
          AND p.status = 'active'
          AND EXISTS (
              SELECT 1 FROM public.workspace_memberships m
              WHERE m.workspace_id = t.workspace_id AND m.user_id = $1
                AND (m.role = 'owner' OR t.created_by = $1)
          )
        RETURNING t.id
    `, userID, workspaceID, projectID, taskID).Scan(&deleted)
    if errors.Is(err, pgx.ErrNoRows) {
        return r.classifyWriteFailure(ctx, userID, workspaceID, projectID, taskID)
    }
    return err
}
