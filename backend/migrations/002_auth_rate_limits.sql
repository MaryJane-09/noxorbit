-- Run once in Supabase SQL Editor before starting the updated backend.
-- Safe for existing users: no existing user accounts or OTPs are deleted.
-- The old code can keep running while this migration is applied.

BEGIN;

ALTER TABLE public.otps
    ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS public.auth_rate_limits (
    key_hash TEXT PRIMARY KEY,
    window_started_at TIMESTAMPTZ NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0)
);

COMMIT;

-- Optional maintenance (run periodically in Supabase SQL Editor or scheduled job):
-- DELETE FROM public.auth_rate_limits
-- WHERE window_started_at < NOW() - INTERVAL '1 day';
