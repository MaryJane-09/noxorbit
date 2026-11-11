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

- Workspace invitations, switching, editing, and deletion
- `GET/POST /api/v1/projects` and project-scoped tasks
- Task updates, comments, activity, notification, and AI APIs

The dashboard currently shows an honest empty state. Its Create Workspace
flow will be connected after the workspace API is tested.
