import { useEffect, useState } from 'react'
import { CalendarDays, CheckCircle2, FolderKanban, LoaderCircle, RefreshCw } from 'lucide-react'
import { changeTaskStatus, fetchMyTasks } from '../api'
import './MyTasksPage.css'

const STATUSES = [
    { value: 'todo', label: 'To Do' },
    { value: 'in_progress', label: 'In Progress' },
    { value: 'done', label: 'Done' },
]
const FILTERS = [
    { value: 'open', label: 'To do & in progress' },
    { value: 'all', label: 'All tasks' },
    { value: 'done', label: 'Completed' },
]

export default function MyTasksPage() {
    const [state, setState] = useState({ items: [], loading: true, error: '' })
    const [reload, setReload] = useState(0)
    const [filter, setFilter] = useState('open')
    const [busyID, setBusyID] = useState(null)
    const [actionError, setActionError] = useState('')

    useEffect(() => {
        const controller = new AbortController()
        fetchMyTasks({ signal: controller.signal })
            .then((items) => {
                if (!controller.signal.aborted) setState({ items, loading: false, error: '' })
            })
            .catch((error) => {
                if (!controller.signal.aborted) setState({ items: [], loading: false, error: error.message })
            })
        return () => controller.abort()
    }, [reload])

    async function updateStatus(task, status) {
        if (busyID || task.status === status || task.project_status !== 'active') return
        setBusyID(task.id)
        setActionError('')
        try {
            const saved = await changeTaskStatus(task.workspace_id, task.project_id, task.id, status)
            setState((current) => ({
                ...current,
                items: current.items.map((item) => item.id === task.id ? { ...item, ...saved } : item),
            }))
        } catch (error) {
            setActionError(error.message || 'Could not update task status.')
        } finally {
            setBusyID(null)
        }
    }

    const displayed = state.items.filter((task) => {
        if (filter === 'done') return task.status === 'done'
        if (filter === 'open') return task.status !== 'done'
        return true
    })
    const completed = state.items.filter((task) => task.status === 'done').length

    return (
        <section className="my-tasks-page" aria-labelledby="my-tasks-title">
            <header className="my-tasks-heading">
                <div>
                    <p className="dashboard-eyebrow">YOUR WORK · ALL WORKSPACES</p>
                    <h1 id="my-tasks-title">My tasks<span>.</span></h1>
                    <p>Every task assigned to you, across the projects and teams you belong to.</p>
                </div>
                <button className="dashboard-feature-action" type="button" disabled={state.loading}
                    onClick={() => { setState((current) => ({ ...current, loading: true, error: '' })); setReload((n) => n + 1) }}>
                    <RefreshCw size={16} aria-hidden="true" /> Refresh
                </button>
            </header>
            {!state.loading && !state.error && <div className="my-tasks-summary">
                <strong>{state.items.length} assigned</strong>
                <span>{state.items.length - completed} remaining</span>
                <span>{completed} completed</span>
            </div>}
            <div className="my-tasks-filters" role="group" aria-label="Filter assigned tasks">
                {FILTERS.map((option) => (
                    <button type="button" key={option.value}
                        className={filter === option.value ? 'is-active' : ''}
                        aria-pressed={filter === option.value}
                        onClick={() => setFilter(option.value)}>{option.label}</button>
                ))}
            </div>
            {actionError && <p role="alert" className="dashboard-form-error">{actionError}</p>}
            {state.loading && <p role="status" className="my-tasks-message"><LoaderCircle size={19} className="dashboard-spin" /> Loading your tasks…</p>}
            {!state.loading && state.error && <div className="my-tasks-message" role="alert">
                <p>{state.error}</p>
                <button type="button" className="dashboard-feature-action" onClick={() => {
                    setState((current) => ({ ...current, loading: true, error: '' }))
                    setReload((n) => n + 1)
                }}>Retry</button>
            </div>}
            {!state.loading && !state.error && displayed.length === 0 && <div className="my-tasks-message">
                <CheckCircle2 size={30} aria-hidden="true" />
                <h2>{state.items.length === 0 ? 'Nothing assigned yet.' : 'No tasks in this view.'}</h2>
                <p>{state.items.length === 0
                    ? 'Once a teammate assigns you a task, it will appear here.'
                    : 'Try another filter to see the rest of your assigned work.'}</p>
            </div>}
            {!state.loading && !state.error && displayed.length > 0 && <div className="my-tasks-list">
                {displayed.map((task) => (
                    <article className="my-tasks-item" key={task.id}>
                        <div className="my-tasks-item-body">
                            <div className="my-tasks-location"><FolderKanban size={15} aria-hidden="true" />
                                <span>{task.workspace_name}</span><span aria-hidden="true">/</span><span>{task.project_name}</span>
                            </div>
                            <h2>{task.title}</h2>
                            {task.description && <p className="my-tasks-description">{task.description}</p>}
                            <div className="my-tasks-meta">
                                <span className={`tasks-priority tasks-priority--${task.priority}`}>{task.priority}</span>
                                {task.due_date && <span><CalendarDays size={15} aria-hidden="true" /> Due {task.due_date}</span>}
                                {task.project_status === 'archived' && <span>Archived · read-only</span>}
                            </div>
                        </div>
                        <div className="my-tasks-control">
                            <label htmlFor={`my-task-status-${task.id}`}>Status</label>
                            <select id={`my-task-status-${task.id}`} value={task.status}
                                disabled={busyID !== null || task.project_status !== 'active'}
                                onChange={(event) => updateStatus(task, event.target.value)}>
                                {STATUSES.map((status) => <option key={status.value} value={status.value}>{status.label}</option>)}
                            </select>
                        </div>
                    </article>
                ))}
            </div>}
        </section>
    )
}
