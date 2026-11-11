-- NoxOrbit: initial workspaces and workspace memberships.
--
-- Run ONCE in the Supabase SQL Editor after confirming public.users(id) is UUID.
-- This migration creates NEW tables only. It does not change existing users,
-- passwords, sessions, OTPs, or authentication data.
--
-- IMPORTANT: NoxOrbit authenticates through its Go backend, not Supabase Auth.
-- These tables intentionally have RLS enabled with NO client-facing policies.
-- The Go backend must use its trusted PostgreSQL connection and enforce
-- per-workspace authorization in every future handler.

BEGIN;

CREATE TABLE public.workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL
        CHECK (char_length(btrim(name)) BETWEEN 1 AND 100),
    purpose TEXT NOT NULL DEFAULT 'general'
        CHECK (purpose IN ('general', 'education', 'development', 'business')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE public.workspace_memberships (
    workspace_id UUID NOT NULL
        REFERENCES public.workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL
        REFERENCES public.users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member'
        CHECK (role IN ('owner', 'member')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (workspace_id, user_id)
);

-- A workspace can have at most one owner. Future Go handlers must also
-- ensure every workspace HAS one owner and cannot remove its last owner.
-- Workspace creation + initial 'owner' membership must share a transaction.
CREATE UNIQUE INDEX workspace_single_owner_idx
    ON public.workspace_memberships (workspace_id)
    WHERE role = 'owner';

-- Supports listing all workspaces visible to an authenticated user.
-- The composite primary key already indexes lookups by workspace_id.
CREATE INDEX workspace_memberships_user_idx
    ON public.workspace_memberships (user_id, workspace_id);

-- Supabase may expose public-schema tables over its Data API. The frontend
-- must NEVER use the database credentials directly. These tables are
-- backend-only: no RLS policies and no anon/authenticated table grants.
ALTER TABLE public.workspaces ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.workspace_memberships ENABLE ROW LEVEL SECURITY;

REVOKE ALL ON TABLE public.workspaces, public.workspace_memberships
    FROM anon, authenticated;

COMMIT;

-- No seed/demo workspaces or memberships are inserted by this migration.
-- If running any statement fails, the transaction rolls back in full.
