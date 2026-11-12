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


## Email-bound invitations (November 2026)

### POST /api/v1/workspaces/{workspace_id}/invitations/email

Requires an authenticated **owner** and `Content-Type: application/json`:

```json
{"email":"teammate@example.com"}
```

The backend validates and normalizes the recipient email, creates a hashed
one-use invitation expiring in seven days, and sends an SMTP message with a
link to `/#invite/<token>`. It returns **201** only after SMTP acknowledges
the message. On delivery failure, it attempts to revoke that invitation and
returns **502**. Invitation secrets are not included in API responses.

To send email invitations, set the Go server environment variable
`FRONTEND_URL` to the **frontend origin**, such as
`http://localhost:5173` for same-device testing or
`https://your-noxorbit-domain.example` after deployment. An HTTPS
origin is required for non-localhost addresses. If missing or invalid,
email invitation requests return **503**. The server's SMTP
credentials must already be configured for verification emails.

Email invitations are **recipient-bound**: acceptance uses the logged-in
user's email from the database, never an email supplied by the browser.
Signing in with another address returns **403**; the intended recipient
must sign in or register and then accept the invitation.

Link and email invitation creation share a workspace-level database quota:
at most **10 invitations per hour** and **30 per day** per workspace.
Quota failures return **429** and are enforced transactionally, including
concurrent requests. Existing invitation history and revocation features
work for either kind of invitation.

**Login continuity:** The frontend preserves the `#invite/<token>`
hash through login and registration, showing the acceptance page after
authentication. An unauthenticated user opening an invitation starts at
the login screen with a route to registration. Email invites sent with a
localhost frontend URL are suitable only for same-device development;
for remote recipients deploy the frontend and set `FRONTEND_URL`
accordingly.

**Deployment note:** No new database migration is needed. The existing
`007_workspace_invitations.sql` schema already supports `kind='email'`.
SMTP and app-origin configuration are required to test actual delivery.


## My Tasks — assigned work across workspaces

`GET /api/v1/my-tasks` requires an authenticated session. The server
takes the user ID only from the validated session and returns tasks where
`assignee_id` equals that user **and** the user has a current membership in
the corresponding workspace. It does not accept an arbitrary `user_id` query
parameter, and it never returns another account's task assignments.

Each result contains the normal task fields plus `project_name`,
`workspace_name`, and `project_status`. Both active and archived project
tasks appear; archived tasks remain read-only. Results are sorted with
unfinished tasks first and earlier due dates before later ones (tasks
without due dates appear last within each group).

The React **My tasks** page (`#my-tasks`) includes filters for unfinished,
completed, and all assignments, and lets the current assignee change a
task's status through the existing protected task-status endpoint.
The backend still checks membership, archived status, and the existing
owner/creator/assignee authorization rules for each update. All other task
editing remains in the project board. The page refreshes on navigation,
manual refresh, and successful local status updates. No new database
migration is needed: it uses tables/indexes from migrations 003/005/006.


## Dashboard Overview cards (UI refinement)

The Overview page now uses existing authenticated API data, without new
backend routes or migrations:
- **Projects:** fetches the current workspace's projects, shows the number
  of active projects, up to two names, and any archived count.
- **My tasks:** fetches the signed-in user's assignments across memberships,
  showing remaining, completed, and total assigned counts.
- **Your upcoming deadlines:** uses the same assigned-task response and
  selects the nearest due date among unfinished assignments; finished tasks
  and tasks without due dates are excluded. This does **not** claim to be
  every deadline in the workspace.

The My Tasks and deadline cards intentionally share a single API call. Data
loads when Overview mounts; switching back from other dashboard screens
refreshes the cards. Errors are displayed rather than replacing them with
fictional data. The duplicated Overview footer was removed and the Projects
cards have more compact spacing with the existing primary Open board action.


## Grouped task deadlines (Overview UI)

The Overview's **Your deadlines** card groups all *unfinished tasks assigned
to the signed-in user* by due calendar date, using the Africa/Lagos date
policy already used to validate task creation. When four assignments are due
tomorrow, the card displays **4 tasks due tomorrow** as a single reminder,
with an expandable group showing all four tasks. Within each date group,
tasks are ranked urgent → high → medium → low, then title and ID for stable
ties. Every due-date group can be expanded independently; the group list is
scrollable so it does not stretch the full Overview unnecessarily.

Warnings are separate from task status and priority. Overdue and due-today
tasks receive red highlights; due-tomorrow and due-soon tasks receive amber.
Completed and undated tasks do not generate deadline warnings. Identical
labels appear in the Task Board and My Tasks views. The Overview calendar
label refreshes approximately every 60 seconds while open.

This is an **in-app dashboard reminder**, not a scheduled background
notification. There is no automatic email digest or push notification yet;
those require a future notification scheduler, delivery infrastructure, and
user notification preferences. The frontend does not generate four emails
for four tasks or modify task priorities or status.

Run `npm run test:deadlines` from `frontend` to verify grouped tomorrow
deadlines, urgency ordering, status filtering, and date boundaries.
No database migration or new API endpoint is required.


## Compact deadline summary and expanded Overview detail panel

The Overview now shows a **compact deadline summary** in its third card:
counts for overdue, due today, due tomorrow, and other upcoming unfinished
assignments. Counts are grouped rather than listing each task in the card.
The `View all deadlines` button toggles a **full-width details panel below
all three cards**. Within that panel each calendar-date group can be expanded
independently, showing task title, priority and workspace/project context.
Date groups remain chronological; task order within each group remains
urgent → high → medium → low. All groups are based on the same authenticated
`GET /api/v1/my-tasks` response as the My Tasks card. There is no extra
request, database migration, email scheduler or automatic status change.

Closing the panel hides its contents without losing the selected expanded
date groups while the Overview remains mounted. At narrower viewport widths
summary cards reflow from three columns to two and then one.


## Task Recycle Bin — migration 008 required

**Deployment order:** Apply `backend/migrations/008_task_recycle_bin.sql` in
Supabase SQL Editor BEFORE running this version of the Go backend. Existing
task queries now refer to `tasks.deleted_at`; without the migration, normal
task listing and updates will fail. No data is removed by the migration.

`DELETE /api/v1/workspaces/{workspace_id}/projects/{project_id}/tasks/{task_id}`
now **soft-deletes** the task, setting `deleted_at`, `deleted_by`, and
`updated_at` rather than removing its row. The existing owner/creator rule
and active-project requirement apply. Normal project-board reads, task-detail
reads and writes, My Tasks, and deadline summaries filter out recycled tasks.

`GET /api/v1/recycle-bin/tasks` lists up to 200 most recently deleted tasks
for which the signed-in member is currently the task creator or a workspace
owner; it includes project and workspace context and timestamps. The user ID
comes exclusively from the authenticated session.

`POST /api/v1/recycle-bin/tasks/{task_id}/restore` restores a recycled task
if its project is active and the signed-in member still has creator/owner
permissions. The guarded SQL UPDATE rechecks those conditions atomically.
It returns 204 when restored, 404 when missing/inaccessible/archived, 401
without a session. The task keeps its ID, assignment, status, and due date.
If the project is archived, restore it in Projects first.

The frontend's `#recycle-bin` route lists deleted tasks and offers Restore.
Existing task delete confirmation now states that deletion is recoverable.

**Scope:** This initial Recycle Bin phase covers tasks only. Projects use the
existing archive/restore flow. Workspace deletion is not enabled.
There is **no automated retention purge**. Authorized users can now
permanently delete individual recycled tasks from the Recycle Bin. Deleted tasks from before migration 008 were permanently
deleted and cannot be recovered by this feature.

Run `go test ./backend/...` and `go vet ./backend/...` from the repository
root; in `frontend`, run `npm run lint` and `npm run build`.
Integration-test against a migrated Supabase instance before production.


## Permanently deleting recycled tasks

`DELETE /api/v1/recycle-bin/tasks/{task_id}` permanently deletes **only**
a task that is already in the recycle bin (`deleted_at IS NOT NULL`).
Its single SQL DELETE rechecks the signed-in user's current workspace
membership and verifies that the user is either a workspace owner or the
task's creator. Deleting a task from a different workspace, guessing a task
ID, or targeting a still-active task cannot bypass those checks. Archived
projects may contain recycled tasks, and those can be permanently removed,
but not restored until the project is active.

The endpoint returns HTTP 204 on success, 404 for a missing or inaccessible
recycled task, and 401 without authentication. No new migration is required
after migration 008.

The Recycle Bin UI displays **Restore task** as the primary action and
**Delete forever** as a secondary destructive action. Delete forever
requires confirmation in an app-owned dialog identifying the task, workspace,
and project, with explicit notice that the action cannot be undone. If the
server rejects the request the task stays in the list and the error is shown.

There is no **Empty bin** bulk action. Deletions are individual, deliberate,
and irreversible.
