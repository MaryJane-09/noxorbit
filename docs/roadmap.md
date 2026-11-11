# NoxOrbit — Product Roadmap

The roadmap is **phased by user value**, not by a fixed calendar. Complete and validate each foundation before introducing the next layer.

See [Product Requirements](prd.md) for detailed requirements, authorization rules, and acceptance criteria.

## Already implemented — Authentication foundation

- Registration and login/logout.
- Email verification and password reset.
- Sessions and initial request/OTP rate limits.
- Authentication-related email notifications.

These existing flows still require production environment security checks and ongoing regression tests.

## Phase 1 — Collaborative workspace MVP

**Audience:** Student groups.

**Outcome:** A team can create a shared workspace, plan a project, assign work, and follow progress.

1. Dashboard shell, workspace empty state, and skippable introduction.
2. Workspaces and membership-aware permissions.
3. Create/switch workspaces; optional purpose preference.
4. Projects within workspaces.
5. Tasks, assignees, deadlines, priority, and To Do / In Progress / Done boards.
6. Real progress counts, personal tasks, and upcoming deadlines.
7. Workspace invitations and cross-workspace authorization tests.

**Do not simulate projects, analytics, or team activity as real data.**

## Phase 2 — Team coordination

**Outcome:** Less work lost in messages and fewer missed updates.

- Comments, mentions, activity history, and notifications.
- Deadline reminders backed by durable scheduling.
- Calendar views/integrations as incremental additions.
- Group discussions/chat, attachments, and voice messages in later increments.
- More flexible roles, workspace rules, and auditability.

## Phase 3 — AI assistance

**Outcome:** Teams understand shared work faster and can act on informed suggestions.

- Project and discussion summaries.
- Blocker and duplicate-task detection.
- Task suggestions and priority recommendations.
- Context-aware workspace assistant.
- Human approval before AI edits; backend-enforced workspace-level AI controls.

## Phase 4 — Specialized workspaces and advanced capabilities

**Outcome:** Extend the validated collaboration core to more kinds of work.

- **Education:** assignment-focused tools, teaching permissions, learning-oriented AI controls.
- **Development:** GitHub integration, PR tracking, CI visibility.
- **Business:** clients, meetings, reports, and optional finance workflows.
- Later possibilities: advanced search, analytics, automation, native mobile apps, offline sync.

A hosted terminal, financial data handling, and offline synchronization require independent security and architecture work before development.

## Immediate next step

Implement **the dashboard shell and first-workspace experience**, then add the workspace backend and database design. Preserve the established NoxOrbit brand and existing authentication flows.
