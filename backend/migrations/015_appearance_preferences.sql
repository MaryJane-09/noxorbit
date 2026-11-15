-- Appearance preferences belong to the account and are independent of
-- calendar preferences. Apply after migration 014.
BEGIN;
CREATE TABLE public.appearance_preferences (
    user_id uuid PRIMARY KEY REFERENCES public.users(id) ON DELETE CASCADE,
    theme text NOT NULL DEFAULT 'dark'
        CHECK (theme IN ('dark', 'light', 'system')),
    reduced_motion boolean NOT NULL DEFAULT false,
    updated_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE public.appearance_preferences ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON public.appearance_preferences FROM anon, authenticated;
COMMIT;
