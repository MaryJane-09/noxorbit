# NoxOrbit API

The Go backend is the only API used by the React frontend. The login session
is stored in an HTTP-only `session` cookie; requests must include credentials.
Unauthenticated requests to workspace endpoints return **401 Unauthorized**.

## Implemented authentication routes (existing)

- `POST /register`
- `POST /verify-otp`
- `POST /login`
- `POST /logout`
- `POST /forgot-password`
- `POST /reset-password`
- `GET /me`
- `GET /health`

## Implemented workspace routes (Phase 1 backend)

### POST /api/v1/workspaces

Create a workspace and its creator's **owner** membership in the same database
transaction. The owner ID comes from the validated session, never the request.

Required header: `Content-Type: application/json`

```json
{
  "name": "Study Group",
  "purpose": "education"
}
```

`name` is required (1–100 characters after trimming). `purpose` is optional:
`general` (default), `education`, `development`, or `business`.
Purpose is a template preference, **not an authorization role**.

Returns **201 Created** and a workspace object:

```json
{
  "id": "a5b4aa87-dc5c-4e29-81c4-b05a0c6179da",
  "name": "Study Group",
  "purpose": "education",
  "role": "owner",
  "created_at": "2026-11-11T12:00:00Z",
  "updated_at": "2026-11-11T12:00:00Z"
}
```

The UUID and timestamps above are illustrative, not live data.
Unknown JSON properties (including `owner_id` and `user_id`) and extra
JSON values are rejected.

### GET /api/v1/workspaces

Returns **200 OK** and a JSON array of **only the current user's
memberships**. For a new account the response is `[]`, not fake data.

### GET /api/v1/workspaces/{id}

Returns **200 OK** and the workspace if the requester belongs to it.
Returns **404 Not Found** both for unknown IDs and for IDs of workspaces
the user cannot access. Malformed UUIDs return **400 Bad Request**.

### PATCH /api/v1/workspaces/{id}

Update a workspace's name and purpose. Only an authenticated workspace
**owner** may edit these settings. The owner permission is checked inside
the PostgreSQL `UPDATE` statement—not just in the browser.

Send `Content-Type: application/json` with both current/new settings:

```json
{
  "name": "Renamed Study Group",
  "purpose": "education"
}
```

Returns **200 OK** with the updated workspace record and new `updated_at`.
Returns **403 Forbidden** for a member without owner permissions and
**404 Not Found** for an unknown or inaccessible workspace. The request
does not accept a client-supplied owner ID, role, or member list.

The existing `003_workspaces.sql` migration already contains the required
columns. **No new migration is necessary.**

### Common behavior

- **401:** Missing or invalid login session.
- **400:** Invalid input or malformed workspace UUID.
- **403:** Logged-in workspace member lacks owner permission for an edit.
- **415:** POST content type is not `application/json`.
- **500:** Database failure; details are only written to backend server logs.

The browser must call these endpoints with `credentials: "include"`.
The backend uses parameterized queries and verifies membership in SQL; the
frontend's visibility of a button or a workspace ID never grants access.

## Planned (not implemented yet)

- Workspace invitations, deletion, and persistent active-workspace preference
- `GET/POST /api/v1/projects` and project-scoped tasks
- Task updates, comments, activity, notification, and AI APIs

## React dashboard integration

The dashboard now has **two separate screens**, selected by hash navigation:

- `#overview` — personalized welcome, compact active workspace, and honest
  project/task/deadline empty states. Before a user has a workspace,
  onboarding still links to creation.
- `#workspaces` — dedicated directory, one **New workspace** action,
  current workspace indicator, `Open overview` switching, and owner-only
  **Edit** settings dialog.

The frontend loads workspaces using `GET /api/v1/workspaces`, creates new
workspaces through `POST /api/v1/workspaces`, and saves name/purpose edits
via `PATCH /api/v1/workspaces/{id}`. The UI updates only after the backend
returns the saved record. Users without owner permission do not see the
Edit action; the Go backend independently enforces the same rule.

Selecting `Open overview` switches the active workspace context in memory.
Refreshing the app selects the newest workspace again, since a persisted
active-workspace preference isn't implemented yet. This UI selection does
not grant API permissions.

If loading fails, the Workspaces screen displays the API error with a
Retry action. Projects, tasks, invitations, and deadlines remain planned;
no sample metrics or fake workspace records are displayed.

### Manual verification

1. Run `go test ./backend/...` and `go vet ./backend/...` from the root.
2. Restart the Go backend, then from `frontend` run `npm run lint`,
   `npm run build`, and `npm run dev`.
3. Sign in and open **Workspaces** in the sidebar. Confirm the current
   Group A / Group B records appear after refreshing.
4. Create a third workspace on the Workspaces page, then refresh; it
   should remain in the list.
5. Use **Open overview** on a different workspace; verify the active
   workspace label and compact Overview summary update.
6. As an owner, select **Edit**, rename the workspace or change purpose,
   and save. Check the saved changes survive a refresh.
7. Confirm Overview no longer contains the full workspace directory or
   the large onboarding illustration after a workspace exists.
8. Verify browser Back/Forward works between `#workspaces` and `#overview`.

For permission verification with a non-owner, use a separate *test*
account with legitimate membership once invitations are supported. The
backend unit tests already cover forbidden and missing-access responses.

**No new Supabase SQL migration is required for this page and edit endpoint.**

## Project Management backend — Phase 1

**Requires:** `backend/migrations/005_projects.sql` applied in Supabase
after workspace migration 003. The React Projects page uses these routes;
before applying migration 005, project requests will fail.

### GET /api/v1/workspaces/{workspace_id}/projects

Lists real projects in a workspace the logged-in user belongs to.
Includes active and archived records. A member with no projects receives
`[]`; users outside the workspace receive **404**.

### POST /api/v1/workspaces/{workspace_id}/projects

Authenticated workspace members can create a project. Example:

```json
{"name":"Group presentation","description":"Plan our slides and review milestones."}
```

Name is required (1–100 characters); description is optional (up to 2000
characters). Workspace ID is taken from the URL and member identity from
the verified session. Returns **201 Created** with the saved record.

### GET /api/v1/workspaces/{workspace_id}/projects/{project_id}

Authenticated members can view a project only within their workspace.
Non-members, unknown projects, and incorrect workspace/project ID pairs
receive **404**, with no cross-workspace data exposed.

### PATCH /api/v1/workspaces/{workspace_id}/projects/{project_id}

Workspace members can edit the project's `name` and `description` using
the same JSON shape as project creation. Unknown JSON fields are rejected.

### PATCH /api/v1/workspaces/{workspace_id}/projects/{project_id}/status

Only owners can archive or restore a project:

```json
{"status":"archived"}
```

Use `{"status":"active"}` to restore. Non-owner members receive **403**;
non-members receive **404**.

Every project is bound to a workspace by database foreign key, and all
repository operations restrict SQL to memberships of that workspace.

### Testing

From the repository root run `go test ./backend/...` and
`go vet ./backend/...`. The GitHub Actions checks workflow also executes
these commands and the frontend lint/build on pushes to `main`.

Password-reset database integration tests run only when explicitly given
a **disposable test database** URL and acknowledgement. See the opt-in
instructions in the security review documentation.

## React Projects page (implemented, requires migration 005)

Selecting **Projects** in the sidebar opens a dedicated `#projects` screen
for the currently selected workspace, without cluttering Overview.

- **New project** creates a real project; it appears only after backend success.
- **Edit** changes a project's name and description for workspace members.
- **Archive/Restore** is shown only to owners and enforced by the Go backend.
- **Active/Archived** tabs show records returned by the API; no fake metrics.
- Overview's Projects card displays the real active and archived project
  counts after loading; errors are shown instead of invented data.
- Changing the active workspace changes the project's data scope.
- Projects continue to exist after refresh because they are in PostgreSQL.

**Manual test:** Apply migration 005, restart the Go backend, pull the
frontend, open Projects, create a project in Group A, switch to Group B,
and verify the project is not displayed there. Edit and archive using the
owner account, then refresh and confirm that changes persist.

Actual project/task deadlines and board features are a separate stage.


## Workspace members and invitation links (new)

### GET /api/v1/workspaces/{workspace_id}/members
Authenticated workspace members can list only their own workspace's members.
Returns `[{ "user_id": "uuid", "name": "Display name", "role": "owner", "joined_at": "timestamp" }]`.
No global user directory or passwords are returned. Nonmembers receive 404.

### POST /api/v1/workspaces/{workspace_id}/invitations
Authenticated owners only. Send `Content-Type: application/json` and `{}`.
Returns 201 with a one-use, seven-day invitation token. The UI wraps the
token in an app link using `#invite/<token>` and lets the owner copy it.
Treat the token as a secret: anyone with it can join the workspace.
The server stores only SHA-256(token), not the raw token.

### POST /api/v1/invitations/accept
Authenticated users only. Send `{ "token": "<invitation-token>" }`.
A valid unused invitation creates a `member` membership in the workspace,
returning `{ "workspace_id": "uuid" }`.
Claims are serialized with a PostgreSQL row lock so two users cannot consume
one use simultaneously. Invalid tokens return 404; expired/used/revoked
tokens return 410. Repeating an acceptance from an existing member is safe.

**Deployment prerequisite:** Apply `007_workspace_invitations.sql` in Supabase
after the workspace migration. Member lists work without 007; link creation
and acceptance require it.

**Not implemented yet:** Sending email invitations, owner revocation UI/API,
removing members, invitation audit UI and rate limits for link generation.
The schema contains `revoked_at` for a future safe revocation workflow.

## Task deadline validation (new)

New tasks and new/changed deadlines reject dates earlier than today's
calendar date in the current MVP's Africa/Lagos timezone (400). Existing
overdue dates may be left unchanged while editing other task details.
The backend also guards SQL mutations against date-validation races.
Existing historic task dates are not automatically changed.


## Workspace management (owner-only)

- `GET /api/v1/workspaces/{workspace_id}/invitations/manage` lists up to 100 latest invitation records with status fields. Original invitation tokens are never returned.
- `DELETE /api/v1/workspaces/{workspace_id}/invitations/{invitation_id}` revokes an active invitation (204). After revocation, any previously shared link is rejected.
- `DELETE /api/v1/workspaces/{workspace_id}/members/{member_id}` removes an ordinary workspace member (204). The owner cannot be removed. Existing task assignments to the removed user become unassigned through the composite foreign key in migration 006; their task creations are preserved.

All these operations require an authenticated workspace owner. They use workspace-scoped database permissions rather than trusting the React UI. Removed users lose access on subsequent authenticated API requests. No new migration is required beyond 007 for these operations.

Invitation email delivery and invitation creation rate limiting are still pending, and generated links must currently be shared privately. Use a disposable test member when checking removal, because removal revokes their workspace access immediately.
