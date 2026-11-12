import { useEffect, useState } from 'react'
import { Activity, ArrowRightLeft, CheckCircle2, ClipboardList, FolderKanban, Layers3, LoaderCircle, RefreshCw, Trash2, Users } from 'lucide-react'
import { fetchWorkspaceActivity } from '../api'
import './ActivityPage.css'

const ACTIONS = {
    'workspace.created': ['created the workspace', Layers3],
    'workspace.updated': ['updated the workspace', Layers3],
    'project.created': ['created a project', FolderKanban],
    'project.updated': ['updated a project', FolderKanban],
    'project.archived': ['archived a project', FolderKanban],
    'project.restored': ['restored a project', FolderKanban],
    'task.created': ['created a task', ClipboardList],
    'task.updated': ['updated a task', ClipboardList],
    'task.assigned': ['changed a task assignment', Users],
    'task.status_changed': ['changed task status', ArrowRightLeft],
    'task.deleted': ['moved a task to the recycle bin', Trash2],
    'task.restored': ['restored a task from the recycle bin', CheckCircle2],
    'task.deleted_forever': ['permanently deleted a task', Trash2],
    'member.joined': ['joined the workspace', Users],
    'member.removed': ['removed a workspace member', Users],
}
const STATUSES = { todo: 'To Do', in_progress: 'In Progress', done: 'Done' }

function eventDetails(event) {
    if (event.action === 'task.status_changed' && event.details?.from && event.details?.to) {
        return `${STATUSES[event.details.from] || event.details.from} → ${STATUSES[event.details.to] || event.details.to}`
    }
    if (event.action === 'workspace.updated' && event.details?.previous_name !== event.entity_name) {
        return `Previously: ${event.details.previous_name}`
    }
    if (event.action === 'project.updated' && event.details?.previous_name !== event.entity_name) {
        return `Previously: ${event.details.previous_name}`
    }
    return null
}

export default function ActivityPage({ activeWorkspace, workspacesLoading, onOpenWorkspaces }) {
    const [state, setState] = useState({ workspaceID: null, items: [], loading: true, error: '' })
    const [reload, setReload] = useState(0)
    const workspaceID = activeWorkspace?.id

    useEffect(() => {
        if (!workspaceID) return
        const controller = new AbortController()
        fetchWorkspaceActivity(workspaceID, { signal: controller.signal })
            .then((items) => {
                if (!controller.signal.aborted) setState({ workspaceID, items, loading: false, error: '' })
            })
            .catch((error) => {
                if (!controller.signal.aborted) setState({ workspaceID, items: [], loading: false, error: error.message })
            })
        return () => controller.abort()
    }, [workspaceID, reload])

    const ready = Boolean(workspaceID && state.workspaceID === workspaceID && !state.loading)
    const items = ready ? state.items : []
    const error = ready ? state.error : ''
    const loading = Boolean(workspaceID) && !ready

    function refresh() {
        setState({ workspaceID: null, items: [], loading: true, error: '' })
        setReload((n) => n + 1)
    }

    return <section className="activity-page" aria-labelledby="activity-page-heading">
        <header className="activity-page-header">
            <div>
                <p className="dashboard-eyebrow">WORKSPACE HISTORY</p>
                <h1 id="activity-page-heading">Activity<span>.</span></h1>
                <p>See what your team changed and when. Events are recorded by Noxorbit, not entered manually.</p>
            </div>
            {workspaceID && <button type="button" className="activity-refresh"
                onClick={refresh} disabled={loading}><RefreshCw size={17} aria-hidden="true" /> Refresh</button>}
        </header>
        <div className="activity-workspace-label">
            <Layers3 size={17} aria-hidden="true" />
            {activeWorkspace ? `Viewing activity in ${activeWorkspace.name}` : 'Select a workspace to view its activity'}
        </div>
        {!workspaceID && <div className="activity-message">
            <Activity size={28} aria-hidden="true" />
            <h2>{workspacesLoading ? 'Loading your workspaces…' : 'No workspace selected.'}</h2>
            <p>Activity is organized by workspace so team histories stay separate.</p>
            {!workspacesLoading && <button type="button" onClick={onOpenWorkspaces}>Go to Workspaces</button>}
        </div>}
        {loading && <div className="activity-message" role="status">
            <LoaderCircle size={22} className="dashboard-spin" aria-hidden="true" />
            Loading workspace activity…
        </div>}
        {error && <div className="activity-message" role="alert">
            <p>{error}</p><button type="button" onClick={refresh}>Try again</button>
        </div>}
        {ready && !error && items.length === 0 && <div className="activity-message">
            <Activity size={30} aria-hidden="true" />
            <h2>No activity recorded yet.</h2>
            <p>New project, task, and workspace updates will appear here after activity tracking is enabled.</p>
        </div>}
        {ready && !error && items.length > 0 && <ol className="activity-timeline">
            {items.map((event) => {
                const [description, Icon] = ACTIONS[event.action] || ['made a change', Activity]
                const extra = eventDetails(event)
                return <li className="activity-event" key={event.id}>
                    <span className="activity-event-icon"><Icon size={19} aria-hidden="true" /></span>
                    <div className="activity-event-content">
                        <p><strong>{event.actor_name || 'Former member'}</strong> {description}</p>
                        <div className="activity-entity-name">{event.entity_name}</div>
                        {extra && <p className="activity-event-detail">{extra}</p>}
                        <time dateTime={event.created_at}>
                            {new Date(event.created_at).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })}
                        </time>
                    </div>
                </li>
            })}
        </ol>}
        {ready && items.length === 100 && <p className="activity-limit-note">Showing the 100 most recent events in this workspace.</p>}
    </section>
}
