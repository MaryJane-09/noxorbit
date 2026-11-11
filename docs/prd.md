# NoxOrbit — Product Requirements Document (PRD)

**Status:** Product direction agreed; feature requirements are planned unless marked implemented.  
**Primary audience for the first release:** Student teams and group projects.  
**Long-term audience:** Developers, businesses, educators, and other collaborative teams.  
**Product name:** NoxOrbit (older drafts may refer to Nexus).

## 1. Product overview

NoxOrbit is an intelligent, adaptable collaborative workspace. It helps teams organize projects, assign responsibilities, track progress, preserve important context, and understand what needs attention.

The initial product is deliberately simple: students should be able to create a workspace, invite teammates, organize group work, and see who owns each task and when it is due.

Over time, NoxOrbit can introduce team communication, AI-powered assistance, and specialized tools for education, software development, and business **without requiring separate accounts or separate applications**.

**Product promise:** Every member should know what needs to be done, who is responsible, what has been completed, and what still needs attention.

## 2. Problem

Teamwork is often fragmented across chat groups, documents, personal notes, and disconnected task lists. This leads to unclear ownership, repeated work, lost decisions, missed deadlines, and time spent trying to understand the current state of a project.

NoxOrbit brings the shared view of work and its context into one understandable place.

## 3. Target users and positioning

**First audience:** Students working on assignments, coursework, club activities, and group projects.

The same core collaboration model should eventually support:
- **Developers:** project planning, GitHub-linked work, pull requests, and CI status.
- **Businesses:** projects, responsibilities, clients, meetings, and reports.
- **Educators:** assignment oversight, appropriate permissions, and learning-oriented AI policies.
- **General teams:** simple task and project collaboration without technical complexity.

Start with students to validate the core workflow. Do not build every specialized experience into the first release.

## 4. Product principles

1. **Clarity over complexity.** Show what matters without forcing unnecessary process.
2. **Context belongs with work.** Keep task descriptions, decisions, and updates connected to their projects.
3. **Ownership is explicit.** Make responsibility and deadlines visible.
4. **Intelligence assists; humans decide.** AI suggestions should be explainable, optional, and reviewable.
5. **Useful defaults; flexible workflows.** Make the first experience simple while leaving space for advanced teams.
6. **Privacy and permissions are foundational.** Access checks must be enforced on the backend.
7. **Performance and accessibility matter.** Keep interfaces responsive, legible, keyboard-accessible, and usable on smaller screens.

## 5. Product terminology and hierarchy

- **Account:** One person's identity. A user may participate in many workspaces.
- **Workspace:** The primary shared environment, whether for a study group, developer team, or business. The user-facing term is **workspace**, not organization.
- **Workspace type:** A preference or starting template (for example, Education, Development, Business, General), **not a different login system or security boundary**.
- **Member:** A user who belongs to a workspace.
- **Project:** A body of work inside a workspace.
- **Task:** An actionable item inside a project, optionally assigned to a workspace member.
- **Board:** A way to view project tasks by status.

An optional organization/enterprise level and teams-within-workspaces can be added later if real user needs justify them.

## 6. Phase 1 — Core collaboration MVP

This phase is successful when a new user can create a workspace, work with teammates, and manage a real project from start to finish.

### 6.1 Authentication and account

- Register with email verification; log in and log out.
- Recover access through a password-reset flow.
- Restore an authenticated session securely when returning to the app.
- View basic profile information.

**Current implementation:** Authentication, email verification, password reset, session handling, and rate-limit protections are already implemented in the Go backend and React frontend. Their production readiness still requires environment-specific validation.

### 6.2 Workspace creation and onboarding

- After first login, show a welcoming **Create your first workspace** experience rather than empty statistics.
- Let the owner name the workspace and choose an optional purpose (Education, Development, Business, General).
- Treat workspace types as presets for future capabilities; the MVP still uses the same underlying core features.
- Offer a brief, skippable walkthrough: create a project, add a task, invite someone, and view the board.
- Returning users can choose among workspaces they belong to.
- A user with no workspaces sees a meaningful empty state, never fabricated demo metrics.

### 6.3 Membership and permissions

- A workspace has an **Owner** and **Members** in the MVP; introduce more granular administrator/editor/viewer permissions when needed.
- The owner can invite others to join. Invitations must be confirmed and limited to the intended workspace.
- Only members can see workspace projects and tasks.
- All reads, edits, assignments, and deletions must verify workspace membership and the relevant permission on the **backend**; hiding a button is not authorization.
- A task may be assigned only to an eligible member of its workspace.
- Do **not** expose a global directory of every user's email address.
- Define safe ownership-transfer and last-owner rules before offering owner deletion or removal.

### 6.4 Projects

- Create a project inside the active workspace.
- List and open projects belonging to that workspace.
- Edit a project's name and description; archive or delete it subject to permission checks.
- Show meaningful project state using real task data.
- Never allow access to a project by guessing its identifier from another workspace.

### 6.5 Tasks and boards

A task in the MVP includes:
- Title (required), optional description.
- Status: **To Do**, **In Progress**, or **Done**.
- Optional assignee, priority, and due date.
- Project and workspace association; creation/update timestamps.

Members with appropriate permissions can create, edit, assign, change status, and delete tasks. A project board groups tasks by status, with a straightforward way to move a task between columns.

Labels, comments, attachments, dependencies, and advanced task templates are **later enhancements**, not prerequisites for the first useful release.

### 6.6 Dashboard

The workspace dashboard shows **real information**, including:
- Projects within the selected workspace.
- Tasks assigned to the current user.
- To Do / In Progress / Done counts.
- Project completion percentage (Done tasks ÷ total tasks × 100; show an empty state rather than a misleading percentage when there are no tasks).
- Upcoming deadlines and overdue work.

A sidebar provides **Overview, Workspaces, Projects, My Tasks**, and a place for future Activity. The top bar provides workspace switching and account controls.

The approved NoxOrbit brand should carry through: **dark charcoal and plum, restrained gold accents, readable typography, spacious layout, and subtle motion**. Do not overload the dashboard with cinematic effects, fake charts, or placeholder metrics.

## 7. Phase 2 — Communication and coordination

Introduce features when Phase 1's core workflow is stable:
- Task comments, mentions, and project activity history.
- Notifications for assignments, relevant changes, and invitations.
- Deadline reminders and optional calendar views/integration.
- Shared project discussions or group chat; file attachments and voice messages as later increments.
- Role and rule improvements: configurable permissions, restrictions, and auditability.

Reminder delivery must use **durable stored jobs or scheduled records** so server restarts do not lose alarms; goroutines alone are not a scheduling system. Notification preferences should prevent unnecessary spam.

## 8. Phase 3 — Workspace intelligence

AI is an **optional assistant to the existing workflow**, not a replacement for teammates.

Potential capabilities:
- Summarize long discussions, meetings, or weekly project activity.
- Explain what changed, what is blocked, and what is due next.
- Suggest tasks from meeting notes or transcripts.
- Identify likely duplicate tasks and recommend a merge or update.
- Recommend personal priorities using visible project context.

**AI safety and control requirements:**
- AI must **not silently merge, delete, assign, or edit work**. Show suggestions and require authorized confirmation before applying changes.
- Workspace owners/authorized administrators may disable NoxOrbit's AI features for learning-focused workspaces.
- These controls apply within NoxOrbit; they cannot block external AI tools.
- Respect workspace permission boundaries when preparing context or displaying AI output.
- Label AI-generated content, explain notable suggestions, and allow users to correct or reject them.
- Define privacy, retention, and third-party model data-sharing policies before connecting any AI provider.

A conversational assistant can be added after reliable core data and permission checks exist.

## 9. Phase 4 — Specialized experiences and scale

All workspace types retain the same account, membership, and permission foundation.

**Education:** Assignment tracking, teacher permissions, group assessment, learning-oriented AI controls.  
**Development:** GitHub repository integration, pull-request status, CI visibility, and developer-centric workflows. A browser-hosted terminal is a distinct high-risk infrastructure feature and is **not** an early requirement.  
**Business:** Client projects, meetings, reports, and optional finance-related workflows.

**Broader future possibilities:** Mobile apps, offline editing with conflict-aware synchronization, semantic search, advanced analytics, and automation. These require separate design and delivery decisions.

## 10. First-time and returning user journeys

### New student team leader
1. Registers or logs in.
2. Creates an Education workspace named for the group.
3. Sees a short, skippable guide explaining projects and tasks.
4. Creates a group project.
5. Adds tasks with priorities and due dates.
6. Invites teammates and assigns responsibilities.
7. Sees accurate progress and deadlines on the dashboard.

### Invited teammate
1. Accepts a workspace invitation and authenticates if required.
2. Sees only the workspace data they are authorized to access.
3. Finds their assigned work and updates task status.
4. Returns later to see current progress.

### Member of multiple workspaces
1. Signs in with one account.
2. Switches between, for example, a student project workspace and a development workspace.
3. Sees different data and, eventually, different workspace-type tools without creating another login.

## 11. Core data and authorization concepts

The initial data model should center on:
- **users** — authenticated accounts (already present).
- **workspaces** — named shared environments, owned by a user.
- **workspace_memberships** — which users belong to a workspace and their role.
- **workspace_invitations** — pending invitations, expiration, and acceptance.
- **projects** — each belongs to exactly one workspace.
- **tasks** — each belongs to a project; assignees must belong to the parent workspace.

A user-to-workspace relationship is **many-to-many through memberships**. Projects must never float outside a workspace. Task access inherits the project's workspace boundary. Prefer explicit membership checks and clear, consistent deletion behavior over relying on client-supplied workspace IDs.

Later phases can introduce comments, activity, notifications, conversations, decisions, milestones, and AI records. Document their relationships before adding migrations.

## 12. Quality and security requirements

- **Access:** Every protected resource endpoint must authenticate the requester and authorize access for that specific workspace or project. Use tests for cross-workspace denial.
- **Data integrity:** Validate task and project inputs; keep ownership and membership relations consistent.
- **Security:** Protect sessions and password-reset flows. Production HTTPS, secure cookies, trusted-proxy configuration, and rate-limit behavior require deployment-specific review.
- **Speed:** Keep common navigation and task changes responsive; measure real-world loading and interaction latency before setting service-level targets. Avoid loading entire workspace histories just to render the overview.
- **Usability:** Responsive desktop/mobile-web layouts, clear empty/error/loading states, keyboard support, and sufficient text contrast.
- **Reliability:** Never claim an operation succeeded until it is persisted. Make reminders and notifications recoverable from server restarts.
- **Honesty of information:** No fake projects, progress bars, activity feeds, or AI insights shown as real user data.

## 13. MVP acceptance criteria

The first release is ready for a student pilot when:
1. A verified user can sign in, sign out, and recover their account.
2. A first-time user can create a workspace and a project.
3. A workspace owner can invite a teammate and that teammate can join.
4. Authorized members can create, assign, edit, and complete tasks.
5. The board correctly reflects task status.
6. Dashboard counts, progress, and deadlines agree with persisted data.
7. Users cannot access another workspace's private data by changing an ID or URL.
8. Empty, loading, and failure states are understandable.
9. The same account can participate in multiple workspaces without mixing data.
10. Existing authentication and email flows continue to pass regression tests.

## 14. Success measures

Start with baselines and pilot feedback rather than invented growth targets:
- Time to create the first workspace and first project.
- Percentage of new users who create their first task.
- Percentage of invited teammates who successfully join.
- Weekly active collaborative workspaces, not just registered accounts.
- Tasks assigned, completed, and overdue.
- Ease of understanding current responsibilities and project status (user feedback).
- Responsiveness and error rates for core user flows.

Use these measures to decide what to improve before expanding into AI and specialized workspaces.

## 15. Explicitly outside the Phase 1 MVP

No full team chat, voice messages, file storage, AI chatbot, automatic task merging, calendar synchronization, finance tools, GitHub terminal, native mobile apps, offline sync, advanced analytics, or organization hierarchy.

These remain part of the **long-term vision**, not abandoned ideas.

## 16. Implementation approach

Build in small, testable slices:

1. Authenticated dashboard shell and first-workspace empty state.
2. Workspace membership model, migrations, and authenticated APIs.
3. Workspace creation and switching.
4. Project creation, listing, and editing.
5. Task CRUD, assignment, and board statuses.
6. Data-driven overview, deadlines, and progress.
7. Invitations, permissions tests, and usability polish.

Every slice should include **backend authorization, error handling, and frontend feedback**. Don't make features appear functional before the corresponding backend behavior exists.

---

**Guiding rule:** NoxOrbit should adapt to how teams work while keeping responsibility, progress, and project context easy to understand.
