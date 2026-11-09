# Authentication security migration (Supabase)

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
