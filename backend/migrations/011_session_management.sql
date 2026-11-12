-- Active sessions for Settings. Apply after migration 010, before updated server.
-- Existing sessions stay valid and get a randomly generated public identifier.
BEGIN;
ALTER TABLE public.sessions
    ADD COLUMN public_id UUID NOT NULL DEFAULT gen_random_uuid(),
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN user_agent TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX sessions_public_id_key ON public.sessions(public_id);
CREATE INDEX sessions_user_expiry_idx ON public.sessions(user_id, expires_at);
COMMIT;
