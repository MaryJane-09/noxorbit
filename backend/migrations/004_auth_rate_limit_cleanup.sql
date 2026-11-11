-- Optional one-time operational migration for NoxOrbit's auth quota table.
-- Requires pg_cron to have been enabled in the Supabase project beforehand.
-- Verify the extension first:
-- SELECT extname FROM pg_extension WHERE extname = 'pg_cron';
-- Do not execute until it appears. Run from Supabase SQL Editor as postgres.
--
-- Cron runs hourly; old 15-minute quota windows are retained for a day.
-- This doesn't alter user accounts, OTPs, or current quota entries.
-- For repeat deployments, first inspect cron.job for an existing job with
-- this name. pg_cron replaces named schedules on supported versions.

SELECT cron.schedule(
    'noxorbit-auth-rate-limit-cleanup',
    '17 * * * *',
    $$DELETE FROM public.auth_rate_limits
      WHERE window_started_at < NOW() - INTERVAL '1 day'$$
);

-- To inspect the schedule:
-- SELECT jobid, jobname, schedule, active FROM cron.job
-- WHERE jobname = 'noxorbit-auth-rate-limit-cleanup';
