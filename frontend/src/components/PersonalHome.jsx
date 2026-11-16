import { useEffect, useState } from 'react'
import { ArrowRight, CalendarClock, CheckCircle2, FolderKanban, Home, Layers3, ListTodo, RefreshCw } from 'lucide-react'
import { fetchMyTasks } from '../api'
import { calendarToday, deadlineLabel, formatCalendarDate } from './deadlines'
import './PersonalHome.css'

function summary(tasks, today) {
    const open = tasks.filter((task) => task.status !== 'done' && task.project_status === 'active')
    return {
        open,
        completed: tasks.filter((task) => task.status === 'done').length,
        dueToday: open.filter((task) => deadlineLabel(task.due_date, today)?.kind === 'today'),
        overdue: open.filter((task) => deadlineLabel(task.due_date, today)?.kind === 'overdue'),
    }
}

// Personal Home includes assignments from every workspace. It is deliberately
// independent of the workspace currently highlighted in the sidebar.
export default function PersonalHome({ user, workspaces, workspacesLoading, workspacesError, preferences, onNavigate, onSelectWorkspace }) {
    const [state, setState] = useState({ tasks: [], loading: true, error: '' })
    const [revision, setRevision] = useState(0)
    useEffect(() => {
        const controller = new AbortController()
        fetchMyTasks({ signal: controller.signal })
            .then((tasks) => { if (!controller.signal.aborted) setState({ tasks, loading: false, error: '' }) })
            .catch((error) => { if (!controller.signal.aborted) setState({ tasks: [], loading: false, error: error.message || 'Could not load tasks.' }) })
        return () => controller.abort()
    }, [revision])

    const today = calendarToday(new Date(), preferences?.time_zone)
    const { open, completed, dueToday, overdue } = summary(state.tasks, today)
    const priorityOrder = { urgent: 0, high: 1, medium: 2, low: 3 }
    const focus = [...open].sort((a, b) => {
        const aDue = a.due_date || '9999-12-31'
        const bDue = b.due_date || '9999-12-31'
        return aDue.localeCompare(bDue) || (priorityOrder[a.priority] ?? 4) - (priorityOrder[b.priority] ?? 4)
    }).slice(0, 4)
    const firstName = (user.name || '').trim().split(/\s+/)[0] || 'there'

    return <div className="personal-home">
        <header className="personal-home-heading">
            <div>
                <p className="dashboard-eyebrow"><Home size={15} aria-hidden="true" /> YOUR HOME · ALL WORKSPACES</p>
                <h1>Welcome back, <span>{firstName}.</span></h1>
                <p>One place to see your priorities and pick up where you left off.</p>
            </div>
            <button className="personal-home-refresh" type="button" disabled={state.loading}
                onClick={() => setRevision((current) => current + 1)}>
                <RefreshCw size={17} aria-hidden="true" /> Refresh
            </button>
        </header>

        {state.error && <div className="personal-home-alert" role="alert">
            <span>{state.error}</span>
            <button type="button" onClick={() => setRevision((current) => current + 1)}>Retry</button>
        </div>}
        <section aria-labelledby="personal-home-glance-heading">
            <div className="personal-home-section-title">
                <div><p className="dashboard-eyebrow">YOUR DAY AT A GLANCE</p>
                    <h2 id="personal-home-glance-heading">What needs your attention</h2></div>
            </div>
            <div className="personal-home-metrics">
                <article className="personal-home-metric">
                    <span className="personal-home-icon"><ListTodo size={21} aria-hidden="true" /></span>
                    <span>Open assignments</span><strong>{state.loading ? '—' : open.length}</strong>
                    <small>Across your active projects</small>
                </article>
                <article className="personal-home-metric">
                    <span className="personal-home-icon"><CalendarClock size={21} aria-hidden="true" /></span>
                    <span>Due today</span><strong>{state.loading ? '—' : dueToday.length}</strong>
                    <small>{state.loading ? 'Checking deadlines' : overdue.length ? `${overdue.length} overdue also need attention` : 'Your calendar, your time zone'}</small>
                </article>
                <article className="personal-home-metric">
                    <span className="personal-home-icon"><CheckCircle2 size={21} aria-hidden="true" /></span>
                    <span>Completed</span><strong>{state.loading ? '—' : completed}</strong>
                    <small>From your assigned tasks</small>
                </article>
            </div>
        </section>

        <div className="personal-home-content">
            <section className="personal-home-panel" aria-labelledby="personal-home-focus-heading">
                <div className="personal-home-panel-heading"><div>
                    <p className="dashboard-eyebrow">KEEP MOVING FORWARD</p>
                    <h2 id="personal-home-focus-heading">Up next</h2>
                    </div>
                    <button className="personal-home-text-action" type="button" onClick={() => onNavigate('my-tasks')}>
                        All my tasks <ArrowRight size={16} aria-hidden="true" /></button>
                </div>
                {state.loading ? <p role="status">Loading your assignments…</p>
                    : focus.length === 0 ? <div className="personal-home-empty">
                        <CheckCircle2 size={24} aria-hidden="true" />
                        <p>No unfinished assignments in active projects right now.</p>
                        <button type="button" onClick={() => onNavigate('my-tasks')}>View my tasks</button>
                    </div>
                        : <div className="personal-home-focus-list">{focus.map((task) =>
                            <div className="personal-home-focus-item" key={task.id}>
                                <span className="personal-home-focus-symbol"><FolderKanban size={18} aria-hidden="true" /></span>
                                <div><strong>{task.title}</strong>
                                    <small>{task.workspace_name} · {task.project_name}</small>
                                    {task.due_date && <small>Due {formatCalendarDate(task.due_date, preferences?.date_format)}</small>}
                                </div>
                                {deadlineLabel(task.due_date, today)?.kind === 'overdue' &&
                                    <span className="personal-home-overdue">Overdue</span>}
                            </div>)}</div>}
            </section>

            <section className="personal-home-panel" aria-labelledby="personal-home-spaces-heading">
                <div className="personal-home-panel-heading"><div>
                    <p className="dashboard-eyebrow">YOUR SPACES</p>
                    <h2 id="personal-home-spaces-heading">My workspaces</h2>
                    </div>
                    <button className="personal-home-text-action" type="button" onClick={() => onNavigate('workspaces')}>
                        View all <ArrowRight size={16} aria-hidden="true" /></button>
                </div>
                {workspacesLoading ? <p role="status">Loading workspaces…</p>
                    : workspacesError ? <p role="alert">{workspacesError}</p>
                        : workspaces.length === 0 ? <div className="personal-home-empty">
                            <p>Create a workspace to collaborate on projects and tasks.</p>
                            <button type="button" onClick={() => onNavigate('workspaces')}>Set up a workspace</button>
                        </div>
                            : <div className="personal-home-space-list">{workspaces.slice(0, 5).map((workspace) =>
                                <button className="personal-home-space" key={workspace.id} type="button"
                                    onClick={() => { onSelectWorkspace(workspace.id); onNavigate('overview') }}>
                                    <span className="personal-home-icon"><Layers3 size={19} aria-hidden="true" /></span>
                                    <span><strong>{workspace.name}</strong><small>{workspace.purpose} · {workspace.role}</small></span>
                                    <ArrowRight size={17} aria-hidden="true" />
                                </button>)}</div>}
            </section>
        </div>
        <p className="personal-home-scope">Home is yours across all workspaces. Choose a workspace to see its separate Overview.</p>
    </div>
}
