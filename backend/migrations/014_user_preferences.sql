-- Settings Phase 2: personal time zone and date format preferences.
-- Apply once after 013. Dates on tasks remain DATE (not timestamps).
BEGIN;
CREATE TABLE public.user_preferences (
 user_id UUID PRIMARY KEY REFERENCES public.users(id) ON DELETE CASCADE,
 time_zone TEXT NOT NULL DEFAULT 'Africa/Lagos'
  CHECK (time_zone IN ('Africa/Lagos','UTC','Europe/London','Europe/Paris',
   'America/New_York','America/Los_Angeles','Asia/Dubai','Asia/Kolkata','Asia/Tokyo')),
 date_format TEXT NOT NULL DEFAULT 'YYYY-MM-DD'
  CHECK (date_format IN ('YYYY-MM-DD','DD/MM/YYYY','MM/DD/YYYY')),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE public.user_preferences ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON public.user_preferences FROM anon,authenticated;
COMMIT;
