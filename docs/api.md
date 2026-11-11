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
