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


## Migration 010 — Profile photos and verified email updates

Run `backend/migrations/010_profile_avatar_email.sql` in Supabase SQL
Editor **after migration 009** and before starting the upgraded Go backend.

This migration adds two optional columns to `public.users` for a bounded
private PNG/JPEG profile photo, and a protected `pending_email_changes`
table for 5-minute email verification codes (stored as bcrypt hashes).
It does not overwrite existing emails, passwords, names, projects or tasks.

Verify:
```sql
SELECT column_name FROM information_schema.columns
WHERE table_schema='public' AND table_name='users'
  AND column_name IN ('avatar_data','avatar_mime');
SELECT to_regclass('public.pending_email_changes');
```
The first query should return two rows, and the second should return
`public.pending_email_changes`.

Avoid testing address changes against your primary account until the
backend and frontend checks have passed.


## Migration 011 — Active account sessions

After migration 010, run `backend/migrations/011_session_management.sql`
in Supabase SQL Editor **before** running the session-management backend.

This adds a random `public_id`, `created_at`, and `user_agent` to
existing `public.sessions` records without invalidating them or exposing
session tokens. New logins include browser user-agent information; older
sessions appear as "Unknown device".

Verify:
```sql
SELECT column_name FROM information_schema.columns
WHERE table_schema='public' AND table_name='sessions'
  AND column_name IN ('public_id','created_at','user_agent')
ORDER BY column_name;
```
Expect three rows. The migration is not intended for repeated execution.

Next, pull the updated code, run the Go tests and vet, then lint/build
the frontend. Test session revocation with two separate browser sessions.


## Migration 012 — Project membership and roles

Apply `backend/migrations/012_project_memberships.sql` **once** in
Supabase SQL Editor after migration 011 and before starting the updated
project/task backend. This migration does **not** erase accounts, projects,
or tasks. It enrolls existing project creators, workspace owners and
existing assignees as project participants, including assignments in the
Recycle Bin. New project creators become project leads automatically.

Project participants must also belong to their workspace. A new task
foreign key prevents assigning tasks to non-participants and clears only
the assignee when participation ends.

Verify:
```sql
SELECT to_regclass('public.project_memberships');
SELECT conname FROM pg_constraint
WHERE conrelid='public.tasks'::regclass
AND conname='tasks_project_assignee_fk';
```
Expect `public.project_memberships` and `tasks_project_assignee_fk`.

Test project membership management and assignments with disposable
workspace accounts. Avoid manually deleting membership records while
testing; use Noxorbit's project-member controls.


## Migration 013 — In-app notifications (Settings Phase 2, first milestone)

Apply `backend/migrations/013_notifications.sql` **once**, after migration
012, in Supabase SQL Editor **before** starting the updated backend. It
creates user-specific notification preferences and notification records,
plus an AFTER INSERT/assignee-change trigger on tasks. Existing tasks,
accounts, and projects are not deleted.

Verify:
```sql
SELECT to_regclass('public.notifications'), to_regclass('public.notification_preferences');
SELECT tgname FROM pg_trigger WHERE tgrelid='public.tasks'::regclass
  AND tgname='noxorbit_task_assignment_notification';
```

**Notifications available now:** other-user task assignments (recorded
when assigned); due tomorrow, due today, and overdue reminders (generated
on a signed-in notification refresh). Repeated visits deduplicate deadline
reminders per task, category and local date; no background scheduler,
push alerts, or email delivery is currently provided. Due dates are
temporarily evaluated using the Africa/Lagos calendar until individual
time zone preferences are implemented.

The notification center appears in the signed-in header. Settings →
Notifications saves all four notification types. Turning a preference off
stops future notifications of that type; previously recorded messages
remain in notification history. The app polls while open approximately
once per minute, and also refreshes when the bell is clicked.

Test with a normal authenticated user after the migration: open the
notification bell, change a preference, reload and confirm persistence;
assign yourself a task with a near deadline and open the bell to see the
appropriate reminder. A separate account is required to test a
non-self assignment notification. The server enforces that only the
authenticated user can read or mark their own notifications.


## Migration 014 — Time zone and date format preferences

Apply `backend/migrations/014_user_preferences.sql` **once** in Supabase
SQL Editor, after 013, **before** starting the updated backend. The API and
notification queries now reference `public.user_preferences`, so they
will fail if the migration is skipped.

Creates an account-scoped `public.user_preferences` table with validated
IANA `time_zone` and `date_format` fields. Existing accounts use
`Africa/Lagos` and `YYYY-MM-DD` defaults without changing their old
task deadlines. Tasks continue to store date-only values (YYYY-MM-DD);
display settings never rewrite or shift those values.

Verify:
```sql
SELECT to_regclass('public.user_preferences');
```
Expect `public.user_preferences`.

Use **Settings → Preferences** to save a time zone and date format.
The chosen time zone changes the calendar day used for due-today,
tomorrow, overdue notifications and creation/editing of future task
dates. Date-format changes affect task, overview, history, notification
and session displays; they do not affect HTML date inputs or API DATE
strings. This release supports nine explicitly listed time zones and
three date formats. Language preferences and email delivery are not yet
implemented.

After pulling changes, run:
```bash
go test ./backend/...
go vet ./backend/...
cd frontend
npm run lint
npm run build
npm run test:deadlines
```
Test saving and reloading the preferences, plus local-midnight behavior.


## Migration 015 — Retired appearance preferences (historical)

Migration `015_appearance_preferences.sql` was part of an experiment
with multiple interface themes. **Noxorbit now uses only its approved
Dark theme.** The Appearance Settings UI and appearance API have been
removed. New installations do **not** need this migration for the
current application.

Do not drop `public.appearance_preferences` automatically on existing
Supabase installations. Historical migrations and already-applied schema
must remain reproducible until a separate, explicitly planned
database-cleanup migration is reviewed. The unused table has no effect
on Noxorbit's active theme, tasks, workspaces, or user preferences.

Operating-system `prefers-reduced-motion` is still respected by the
signed-in dashboard.
