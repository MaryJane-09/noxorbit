-- NoxOrbit workspace invitation foundation.
-- Apply once after 003_workspaces. No existing records are changed.
-- Tokens must be created by crypto/rand, delivered securely, and ONLY
-- their SHA-256 hashes may be stored. This table alone does not send invites.
BEGIN;
CREATE TABLE public.workspace_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES public.workspaces(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES public.users(id),
    kind TEXT NOT NULL CHECK (kind IN ('email','link')),
    invited_email TEXT,
    token_hash BYTEA NOT NULL UNIQUE CHECK (octet_length(token_hash)=32),
    expires_at TIMESTAMPTZ NOT NULL,
    max_uses INTEGER NOT NULL DEFAULT 1 CHECK (max_uses BETWEEN 1 AND 100),
    use_count INTEGER NOT NULL DEFAULT 0 CHECK (use_count >= 0 AND use_count <= max_uses),
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        (kind = 'email' AND invited_email IS NOT NULL AND length(btrim(invited_email)) > 3 AND max_uses = 1)
        OR (kind = 'link' AND invited_email IS NULL)
    ),
    CHECK (expires_at > created_at)
);
CREATE INDEX workspace_invitations_active_idx ON public.workspace_invitations(workspace_id,expires_at)
    WHERE revoked_at IS NULL;
ALTER TABLE public.workspace_invitations ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON public.workspace_invitations FROM anon, authenticated;
COMMIT;
