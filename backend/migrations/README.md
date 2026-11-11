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

**Status:** Apply manually in the Supabase SQL Editor before implementing
workspace repositories and HTTP routes. No application code currently uses
these new tables.

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
