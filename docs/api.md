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

### Common behavior

- **401:** Missing or invalid login session.
- **400:** Invalid input or malformed workspace UUID.
- **415:** POST content type is not `application/json`.
- **500:** Database failure; details are only written to backend server logs.

The browser must call these endpoints with `credentials: "include"`.
The backend uses parameterized queries and verifies membership in SQL; the
frontend's visibility of a button or a workspace ID never grants access.

## Planned (not implemented yet)

- Workspace invitations, editing, deletion, and persistent active-workspace preference
- `GET/POST /api/v1/projects` and project-scoped tasks
- Task updates, comments, activity, notification, and AI APIs

## React dashboard integration

The dashboard now loads the signed-in user's workspaces using
`GET /api/v1/workspaces`. The "Create workspace" dialog sends the validated
name and selected purpose to `POST /api/v1/workspaces`. A workspace is
displayed only after the backend returns the saved record.

Users can choose which of **their loaded workspaces** to view from the
dashboard. The current selection is kept in React state, not the database;
reloading selects the newest workspace again. This is navigation context,
not a new authorization grant. Every future project/task endpoint must
still check the session's workspace membership independently.

If the API is unavailable, the dashboard displays an error with a Retry
button instead of claiming the account has no workspaces. The projects,
tasks, invitations, and deadlines features are **not implemented** and
continue to show explanatory placeholders, never fabricated data.

### Manual verification

1. Run `go test ./backend/...` and `go vet ./backend/...` from the root.
2. Start the Go backend, then start the React frontend with `npm run dev`.
3. Log in, select **Create workspace**, give it a name and purpose, and submit.
4. Confirm it appears under **Your workspaces** and as the current workspace.
5. Refresh; the workspace should still be listed because it is in PostgreSQL.
6. Create a second workspace and check you can switch between them.
7. Optionally verify new rows in Supabase's `workspaces` and
   `workspace_memberships` tables, without sharing personal login data.

The frontend should also pass `npm run lint` and `npm run build`.

