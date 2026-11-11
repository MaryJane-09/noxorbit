# NoxOrbit — Database Design

**Backend:** Go with a direct, trusted PostgreSQL connection to Supabase.  
**Client:** React talks to Go HTTP endpoints; it does **not** query
PostgreSQL or Supabase Data API tables directly.

## Current authentication tables

The Go authentication system already uses existing database records for
accounts, verification codes, sessions, and request quotas. Do not recreate
those tables through the workspace migration.

- `users` — existing accounts; `id` is a UUID.
- `sessions` — existing login sessions.
- `otps` — existing email verification and password-reset codes.
- `auth_rate_limits` — existing rate-limit state.

## Phase 1: workspaces and memberships

Introduced by **`backend/migrations/003_workspaces.sql`**.

### `workspaces`

| Column | Meaning |
| --- | --- |
| `id` | UUID primary key; PostgreSQL generates it when omitted. |
| `name` | Required workspace name, 1–100 non-whitespace characters. |
| `purpose` | `general`, `education`, `development`, or `business`. A preference, not a login type or access boundary. |
| `created_at` | Creation timestamp. |
| `updated_at` | Timestamp to be updated when future API handlers modify the workspace. |

### `workspace_memberships`

| Column | Meaning |
| --- | --- |
| `workspace_id` | UUID foreign key to `workspaces.id`. |
| `user_id` | UUID foreign key to existing `users.id`. |
| `role` | `owner` or `member`. |
| `joined_at` | Time the user joined the workspace. |

`(workspace_id, user_id)` is the composite primary key, preventing duplicate
memberships. A partial unique index permits **at most one owner** per workspace.

**Relationship:** Users ⇄ workspace memberships ⇄ workspaces.
A user may belong to several workspaces, and a workspace may contain many
users. The original account is never duplicated for each workspace.

### Important invariants for the Go backend

- **Creation is transactional:** Insert a workspace and its creator's
  `owner` membership inside **one database transaction**. If either fails,
  roll back both operations.
- **At least one owner:** The unique index guarantees *at most one owner*,
  but it does not enforce *at least one*. Until ownership transfer is
  implemented safely, prohibit owner removal and role demotion.
- **Authorization:** Obtain the requesting user ID from their validated
  session, not the request body. Every workspace read or mutation must
  check membership; owner-only actions must check `role = 'owner'`.
- **Privacy:** Never return every workspace or every user's email address.
  List only workspaces to which the requester belongs.
- **Lifecycle:** Deleting a workspace cascades to memberships; user deletion
  cascades to their memberships. Application rules must prevent deleting
  a workspace's last owner without an explicit workspace deletion flow.
- **Future tables:** Projects must reference a workspace; tasks must reference
  projects; assignments must respect workspace membership.

### Supabase / Row Level Security

The migration enables RLS on both new public tables and creates **no
anon/authenticated policies**. It also revokes the standard Supabase Data
API roles' direct table privileges. This is intentional: the Go backend
is the API and owns the authorization logic.

The trusted Go database connection needs permission to operate despite
these RLS rules (the standard Supabase `postgres` role ordinarily has that
capability). Do not use client-side Supabase keys to bypass the backend,
and do not embed a privileged database connection string in React.

When adding a differently privileged PostgreSQL role or deploying to a new
environment, verify database privileges and RLS behavior before enabling
workspace APIs. **RLS does not replace Go handler authorization.**

## Planned later (not created by migration 003)

- `workspace_invitations` — token-based invitations with expiry and acceptance.
- `projects` — workspace-owned projects.
- `tasks` — project tasks, statuses, assignees, and deadlines.
- `comments`, `activity`, `notifications`, `decisions`,
  `milestones` — collaboration and intelligence phases.

See [Product Requirements](prd.md) for feature boundaries.
