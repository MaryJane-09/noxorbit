# NoxOrbit Database Migrations (Supabase)

The NoxOrbit backend uses Supabase PostgreSQL to keep OTP attempt counts and
request quotas consistent across concurrent requests and server restarts.

## Apply BEFORE starting the updated backend

1. Open the **NoxOrbit project** in Supabase > **SQL Editor**.
2. Review and execute `backend/migrations/002_auth_rate_limits.sql`.
3. Confirm the new column and table:

```sql
SELECT column_name FROM information_schema.columns
WHERE table_schema = 'public' AND table_name = 'otps'
  AND column_name = 'attempts';

SELECT to_regclass('public.auth_rate_limits');
```

The result should contain `attempts` and `auth_rate_limits`, respectively.
Migration 002 does **not** delete users, OTPs or passwords. The old running
backend does not use the new column and can continue running during migration.

Then pull the updated code, run `go test ./backend/...`,
`go vet ./backend/...`, and restart the backend.

## How the limits work

- Each registration or password-reset code is valid for 5 minutes.
- The fifth incorrect guess locks that code. A correct match consumes it.
- Codes may be reissued for the same email and purpose once per 60 seconds.
- Login attempts are limited to 12 per email per 15 minutes, whether login
  succeeds or fails. The email is hashed before becoming a database key.
- Additional per-TCP-peer IP quotas last 15 minutes: registration 12,
  login 60, email-reset requests 12, verification 40, password reset 40.
- An exceeded quota returns HTTP 429. A rate-limit database failure returns
  HTTP 503 rather than allowing unbounded attempts.

**Deployment note:** The IP limiter currently uses the direct TCP peer
(`RemoteAddr`), ignoring untrusted `X-Forwarded-For` headers. If you deploy
behind a reverse proxy or shared gateway, several real users may share that
peer IP and therefore the same quota. Configure a trusted-proxy strategy
before public production deployment. This is a starter safeguard, not a
complete anti-abuse system.

## Housekeeping

Rate-limit entries are pseudonymous hashes; entries for abandoned keys can
be deleted after their fixed windows have ended. Run this occasionally:

```sql
DELETE FROM public.auth_rate_limits
WHERE window_started_at < NOW() - INTERVAL '1 day';
```

Use a development or test account when checking failed attempts. Never test
against someone else's account or share real login credentials or OTP codes.

## Migration 003 — Workspaces and memberships (Phase 1)

**Status:** Implemented. The Go workspace API depends on these tables;
apply this migration before starting the workspace-enabled backend.

1. Open the same Supabase project your Go backend uses.
2. Optionally confirm `public.users.id` is a UUID:

```sql
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_schema = 'public'
  AND table_name = 'users'
  AND column_name = 'id';
```

The expected `data_type` is `uuid`. If not, **do not run migration 003**
until the mismatch is investigated.

3. Open and review `backend/migrations/003_workspaces.sql`, then run the
   full file **once** in a new Supabase SQL Editor query.
4. Verify that both tables and their key columns exist:

```sql
SELECT table_name, column_name, data_type
FROM information_schema.columns
WHERE table_schema = 'public'
  AND table_name IN ('workspaces', 'workspace_memberships')
ORDER BY table_name, ordinal_position;

SELECT tablename, rowsecurity
FROM pg_tables
WHERE schemaname = 'public'
  AND tablename IN ('workspaces', 'workspace_memberships');
```

Expect `workspaces` and `workspace_memberships`, with `rowsecurity = true`
for both. Neither table should contain records immediately after creation.

**Why no `IF NOT EXISTS`?** This is a one-time schema change. If either
table already exists, the migration should fail clearly instead of silently
accepting an incompatible structure. The surrounding `BEGIN`/`COMMIT`
keeps the two-table creation atomic.

**Security note:** The migration blocks direct access through Supabase's
standard anonymous/authenticated Data API roles using RLS without policies
and revoked privileges. The backend uses its own trusted PostgreSQL
connection, and future Go handlers must independently enforce workspace
membership and ownership. This migration does not create any frontend
workspace UI, HTTP routes, or workspace records.

**Do not** run migration 003 against a production database without first
reviewing it and confirming the expected existing `users` schema.

## Migration 001 — Legacy password-data cleanup (review first)

The previous pre-verification `pendings` table held plaintext passwords.
The current registration flow no longer uses it. Check whether it remains:

```sql
SELECT to_regclass('public.pendings');
```

If this returns `pendings`, review and run the **existing**
`backend/migrations/001_drop_legacy_pendings.sql` in Supabase SQL Editor.
**This permanently deletes unfinished legacy registrations.** Do not
export, share, or keep copies of the plaintext pending-password data.
If it returns `NULL`, there is nothing to remove.

## Migration 004 — Scheduled rate-limit housekeeping (optional)

The `auth_rate_limits` table will accumulate stale hashed quota keys
until cleaned up. The backend currently reads only fresh windows, but
old rows still occupy storage.

1. In Supabase, enable the `pg_cron` extension using the supported
   Database Extensions/Cron feature for your project.
2. Verify installation via
   `SELECT extname FROM pg_extension WHERE extname = 'pg_cron';`.
3. Review and run `004_auth_rate_limit_cleanup.sql` **once**.
4. Check the scheduled job in `cron.job`.

This is an operational job, not an application schema prerequisite. If
you cannot enable Cron in your project, run the documented delete query
periodically through a trusted administrative connection instead.

## Migration 005 — Projects (Phase 1)

Run **after migration 003** and before enabling the project backend routes:

1. Review and run `backend/migrations/005_projects.sql` **once** in
   Supabase SQL Editor. Migration 004 is optional; it need not run first.
2. Verify the table and RLS settings:

```sql
SELECT to_regclass('public.projects');
SELECT tablename, rowsecurity FROM pg_tables
WHERE schemaname = 'public' AND tablename = 'projects';
```

Expected: `public.projects` exists and `rowsecurity` is `true`.
No dummy projects are inserted. The backend enforces workspace membership
in each project query and owner permission for archive/restore.
