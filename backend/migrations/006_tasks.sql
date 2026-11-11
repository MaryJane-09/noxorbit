-- NoxOrbit Task Management — migration 006.
-- Apply in Supabase SQL Editor AFTER migrations 003_workspaces and 005_projects.
-- Run once. This adds task data only; it does not change existing accounts,
-- authentication, workspaces, memberships, or existing project records.
--
-- NoxOrbit authenticates via its Go API, NOT Supabase Auth. RLS is enabled,
-- but task-level authorization must be implemented by the Go API.
-- PostgreSQL 15+ is required for SET NULL (assignee_id) on the composite FK.

BEGIN;

-- Ensure a task's project ID and workspace ID always refer to the SAME project.
-- The existing projects.id primary key remains unchanged.
ALTER TABLE public.projects
    ADD CONSTRAINT projects_id_workspace_unique UNIQUE (id, workspace_id);

CREATE TABLE public.tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    workspace_id UUID NOT NULL,
    title TEXT NOT NULL
        CHECK (char_length(btrim(title)) BETWEEN 1 AND 200),
    description TEXT NOT NULL DEFAULT ''
        CHECK (char_length(description) <= 5000),
    status TEXT NOT NULL DEFAULT 'todo'
        CHECK (status IN ('todo', 'in_progress', 'done')),
    priority TEXT NOT NULL DEFAULT 'medium'
        CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    due_date DATE,
    created_by UUID
        REFERENCES public.users(id) ON DELETE SET NULL,
    assignee_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT tasks_project_workspace_fk
        FOREIGN KEY (project_id, workspace_id)
        REFERENCES public.projects(id, workspace_id)
        ON DELETE CASCADE,

    -- A task can be unassigned. If an assigned member leaves the workspace,
    -- clear only assignee_id, retaining the task's workspace_id.
    CONSTRAINT tasks_workspace_assignee_fk
        FOREIGN KEY (workspace_id, assignee_id)
        REFERENCES public.workspace_memberships(workspace_id, user_id)
        ON DELETE SET NULL (assignee_id)
);

-- Listing tasks by project and Kanban status.
CREATE INDEX tasks_project_status_created_idx
    ON public.tasks (project_id, status, created_at DESC, id DESC);

-- My Tasks and upcoming deadlines within a workspace.
CREATE INDEX tasks_workspace_assignee_due_idx
    ON public.tasks (workspace_id, assignee_id, due_date)
    WHERE assignee_id IS NOT NULL;

-- Project deadline views; null due dates naturally sort last with ASC.
CREATE INDEX tasks_project_due_idx
    ON public.tasks (project_id, due_date)
    WHERE due_date IS NOT NULL;

-- Frontend must use the trusted Go API. Direct Supabase Data API access to
-- tasks is disallowed; the API must enforce authenticated workspace membership
-- for reads and the agreed owner/creator/assignee permissions for writes.
ALTER TABLE public.tasks ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON TABLE public.tasks FROM anon, authenticated;

COMMIT;

-- No seed tasks, fabricated task counts, or assignment-acceptance statuses.
-- Future assignment acceptance and recycle-bin features get separate migrations.
