-- Noxorbit Settings: avatars and verified email change.
-- Run AFTER migration 009. Do not run against a non-test database without a backup.
BEGIN;
ALTER TABLE public.users
    ADD COLUMN avatar_data BYTEA,
    ADD COLUMN avatar_mime TEXT CHECK (avatar_mime IN ('image/png', 'image/jpeg')),
    ADD CONSTRAINT users_avatar_pair CHECK ((avatar_data IS NULL) = (avatar_mime IS NULL)),
    ADD CONSTRAINT users_avatar_size CHECK (avatar_data IS NULL OR octet_length(avatar_data) BETWEEN 1 AND 524288);

CREATE TABLE public.pending_email_changes (
    user_id UUID PRIMARY KEY REFERENCES public.users(id) ON DELETE CASCADE,
    new_email TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    attempts SMALLINT NOT NULL DEFAULT 0 CHECK (attempts BETWEEN 0 AND 5),
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    CHECK (length(new_email) BETWEEN 3 AND 254)
);
CREATE INDEX pending_email_changes_expiry_idx ON public.pending_email_changes(expires_at);
ALTER TABLE public.pending_email_changes ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON public.pending_email_changes FROM anon,authenticated;
COMMIT;
