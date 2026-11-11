-- NoxOrbit Project Management — Phase 1, migration 005.
-- Run ONCE in Supabase SQL Editor only AFTER migration 003 workspaces
-- has been applied. Migration 004 is optional maintenance; not a prerequisite.
-- Existing accounts, workspaces, and memberships are not changed.

BEGIN;

CREATE TABLE public.projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL
        REFERENCES public.workspaces(id) ON DELETE CASCADE,
    name TEXT NOT NULL
        CHECK (char_length(btrim(name)) BETWEEN 1 AND 100),
    description TEXT NOT NULL DEFAULT ''
        CHECK (char_length(description) <= 2000),
    status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'archived')),
    created_by UUID
        REFERENCES public.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX projects_workspace_created_idx
    ON public.projects (workspace_id, created_at DESC, id DESC);

-- The frontend calls the trusted Go API. The Supabase Data API must not
-- expose projects directly; authorization lives in workspace-scoped Go SQL.
ALTER TABLE public.projects ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON TABLE public.projects FROM anon, authenticated;

COMMIT;

-- No seed projects or placeholder task counts.
