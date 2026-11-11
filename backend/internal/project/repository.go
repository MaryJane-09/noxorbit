package project

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// A user must be a member of the route's workspace to see its projects.
// Missing/inaccessible workspaces are intentionally indistinguishable.
func (r *Repository) isMember(ctx context.Context, userID, workspaceID uuid.UUID) (bool, error) {
	var member bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1 FROM public.workspace_memberships
		WHERE user_id = $1 AND workspace_id = $2
	)`, userID, workspaceID).Scan(&member)
	return member, err
}

func scanProject(row pgx.Row) (Project, error) {
	var p Project
	err := row.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *Repository) Create(ctx context.Context, userID, workspaceID uuid.UUID, input Input) (Project, error) {
	p, err := scanProject(r.pool.QueryRow(ctx, `
		INSERT INTO public.projects (workspace_id, name, description, created_by)
		SELECT $1, $2, $3, $4
		WHERE EXISTS (
			SELECT 1 FROM public.workspace_memberships
			WHERE workspace_id = $1 AND user_id = $4
		)
		RETURNING id, workspace_id, name, description, status, created_at, updated_at
	`, workspaceID, input.Name, input.Description, userID))
	if errors.Is(err, pgx.ErrNoRows) { return Project{}, ErrWorkspaceNotFound }
	return p, err
}

func (r *Repository) ListForUser(ctx context.Context, userID, workspaceID uuid.UUID) ([]Project, error) {
	// Check access even when no projects exist.
	member, err := r.isMember(ctx, userID, workspaceID)
	if err != nil { return nil, err }
	if !member { return nil, ErrWorkspaceNotFound }

	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.workspace_id, p.name, p.description, p.status, p.created_at, p.updated_at
		FROM public.projects p
		WHERE p.workspace_id = $1
		  AND EXISTS (
			SELECT 1 FROM public.workspace_memberships m
			WHERE m.workspace_id = p.workspace_id AND m.user_id = $2
		  )
		ORDER BY p.created_at DESC, p.id DESC
	`, workspaceID, userID)
	if err != nil { return nil, err }
	defer rows.Close()

	items := make([]Project, 0)
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *Repository) GetForUser(ctx context.Context, userID, workspaceID, projectID uuid.UUID) (Project, error) {
	p, err := scanProject(r.pool.QueryRow(ctx, `
		SELECT p.id, p.workspace_id, p.name, p.description, p.status, p.created_at, p.updated_at
		FROM public.projects p
		WHERE p.id = $1 AND p.workspace_id = $2
		  AND EXISTS (
			SELECT 1 FROM public.workspace_memberships m
			WHERE m.workspace_id = p.workspace_id AND m.user_id = $3
		  )
	`, projectID, workspaceID, userID))
	if errors.Is(err, pgx.ErrNoRows) { return Project{}, ErrNotFound }
	return p, err
}

// Members can edit project details. The same membership constraint that
// guards reads is part of the UPDATE; guessing an ID grants nothing.
func (r *Repository) UpdateForMember(ctx context.Context, userID, workspaceID, projectID uuid.UUID, input Input) (Project, error) {
	p, err := scanProject(r.pool.QueryRow(ctx, `
		UPDATE public.projects p
		SET name = $1, description = $2, updated_at = NOW()
		WHERE p.id = $3 AND p.workspace_id = $4
		  AND EXISTS (
			SELECT 1 FROM public.workspace_memberships m
			WHERE m.workspace_id = p.workspace_id AND m.user_id = $5
		  )
		RETURNING id, workspace_id, name, description, status, created_at, updated_at
	`, input.Name, input.Description, projectID, workspaceID, userID))
	if errors.Is(err, pgx.ErrNoRows) { return Project{}, ErrNotFound }
	return p, err
}

// Only owners may archive or restore a project; this is enforced in SQL.
func (r *Repository) UpdateStatusForOwner(ctx context.Context, userID, workspaceID, projectID uuid.UUID, status string) (Project, error) {
	p, err := scanProject(r.pool.QueryRow(ctx, `
		UPDATE public.projects p
		SET status = $1, updated_at = NOW()
		WHERE p.id = $2 AND p.workspace_id = $3
		  AND EXISTS (
			SELECT 1 FROM public.workspace_memberships m
			WHERE m.workspace_id = p.workspace_id AND m.user_id = $4 AND m.role = 'owner'
		  )
		RETURNING id, workspace_id, name, description, status, created_at, updated_at
	`, status, projectID, workspaceID, userID))
	if err == nil { return p, nil }
	if !errors.Is(err, pgx.ErrNoRows) { return Project{}, err }
	visible, getErr := r.GetForUser(ctx, userID, workspaceID, projectID)
	if getErr != nil { return Project{}, getErr }
	_ = visible
	return Project{}, ErrForbidden
}
