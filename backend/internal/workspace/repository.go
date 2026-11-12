package workspace

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

// Create saves the workspace and the creator's owner membership atomically.
// If either INSERT (or the COMMIT) fails, the workspace is not created.
func (r *Repository) Create(ctx context.Context, ownerID uuid.UUID, name, purpose string) (Workspace, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Workspace{}, err
	}
	defer tx.Rollback(ctx) // Safe after Commit; errors are intentionally ignored.

	var w Workspace
	err = tx.QueryRow(ctx, `
		INSERT INTO public.workspaces (name, purpose, activity_actor_id)
		VALUES ($1, $2, $3)
		RETURNING id, name, purpose, created_at, updated_at
	`, name, purpose, ownerID).Scan(&w.ID, &w.Name, &w.Purpose, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return Workspace{}, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO public.workspace_memberships (workspace_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, w.ID, ownerID)
	if err != nil {
		return Workspace{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Workspace{}, err
	}
	w.Role = "owner"
	return w, nil
}

// ListForUser only returns workspaces joined to the authenticated user's ID.
// Querying public.workspaces without this JOIN would expose other teams.
func (r *Repository) ListForUser(ctx context.Context, userID uuid.UUID) ([]Workspace, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT w.id, w.name, w.purpose, m.role, w.created_at, w.updated_at
		FROM public.workspaces AS w
		JOIN public.workspace_memberships AS m ON m.workspace_id = w.id
		WHERE m.user_id = $1
		ORDER BY w.created_at DESC, w.id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workspaces := make([]Workspace, 0)
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.Purpose, &w.Role, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, w)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return workspaces, nil
}

// GetForUser deliberately uses the same "not found" response for both an
// unknown workspace and one the current user is not allowed to access.
func (r *Repository) GetForUser(ctx context.Context, userID, workspaceID uuid.UUID) (Workspace, error) {
	var w Workspace
	err := r.pool.QueryRow(ctx, `
		SELECT w.id, w.name, w.purpose, m.role, w.created_at, w.updated_at
		FROM public.workspaces AS w
		JOIN public.workspace_memberships AS m ON m.workspace_id = w.id
		WHERE w.id = $1 AND m.user_id = $2
	`, workspaceID, userID).Scan(&w.ID, &w.Name, &w.Purpose, &w.Role, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Workspace{}, ErrNotFound
	}
	if err != nil {
		return Workspace{}, err
	}
	return w, nil
}

// UpdateForOwner changes only workspaces for which the authenticated user
// has an owner membership. The authorization check is part of the UPDATE
// statement, so a client-provided workspace ID alone never grants editing.
func (r *Repository) UpdateForOwner(ctx context.Context, userID, workspaceID uuid.UUID, name, purpose string) (Workspace, error) {
	var w Workspace
	err := r.pool.QueryRow(ctx, `
		UPDATE public.workspaces AS w
		SET name = $1, purpose = $2, activity_actor_id = $4, updated_at = NOW()
		WHERE w.id = $3
		  AND EXISTS (
			SELECT 1 FROM public.workspace_memberships AS m
			WHERE m.workspace_id = w.id AND m.user_id = $4 AND m.role = 'owner'
		  )
		RETURNING w.id, w.name, w.purpose, w.created_at, w.updated_at
	`, name, purpose, workspaceID, userID).Scan(&w.ID, &w.Name, &w.Purpose, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		// Avoid exposing workspaces to users who are not members. Members
		// without the owner role receive a clear permission error.
		visible, lookupErr := r.GetForUser(ctx, userID, workspaceID)
		if lookupErr != nil {
			return Workspace{}, lookupErr
		}
		if visible.Role != "owner" {
			return Workspace{}, ErrForbidden
		}
		return Workspace{}, ErrNotFound
	}
	if err != nil {
		return Workspace{}, err
	}
	w.Role = "owner"
	return w, nil
}
