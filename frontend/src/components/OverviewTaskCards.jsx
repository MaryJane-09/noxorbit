import { useEffect, useState } from 'react'
import { ArrowRight, CalendarDays, CheckCircle2, ChevronDown, ListTodo } from 'lucide-react'
import { fetchMyTasks } from '../api'
import { calendarToday, groupAssignedDeadlines } from './deadlines'

// The two summary cards and expandable panel share a single authenticated request.
export default function OverviewTaskCards({ onOpenTasks }) {
    const [state, setState] = useState({ items: [], loading: true, error: '', today: '' })
    const [detailsOpen, setDetailsOpen] = useState(false)
    const [expandedDates, setExpandedDates] = useState([])

    useEffect(() => {
        const controller = new AbortController()
        fetchMyTasks({ signal: controller.signal })
            .then((items) => {
                if (!controller.signal.aborted) setState({ items, loading: false, error: '', today: calendarToday() })
            })
            .catch((error) => {
                if (!controller.signal.aborted) setState({
                    items: [], loading: false,
                    error: error.message || 'Could not load your tasks.', today: calendarToday(),
                })
            })
        const interval = window.setInterval(() => {
            if (!controller.signal.aborted) setState((current) => ({ ...current, today: calendarToday() }))
        }, 60_000)
        return () => { controller.abort(); window.clearInterval(interval) }
    }, [])

    const unfinished = state.items.filter((task) => task.status !== 'done')
    const completed = state.items.length - unfinished.length
    const groups = groupAssignedDeadlines(state.items, state.today)
    const countByTiming = (kind) => groups
        .filter((group) => group.timing.kind === kind)
        .reduce((sum, group) => sum + group.items.length, 0)
    const overdue = countByTiming('overdue')
    const today = countByTiming('today')
    const tomorrow = countByTiming('tomorrow')
    const upcoming = countByTiming('soon') + countByTiming('upcoming')
    const datedCount = overdue + today + tomorrow + upcoming
    const reminders = [
        { kind: 'overdue', count: overdue, label: 'overdue' },
        { kind: 'today', count: today, label: 'due today' },
        { kind: 'tomorrow', count: tomorrow, label: 'due tomorrow' },
        { kind: 'upcoming', count: upcoming, label: 'upcoming' },
    ].filter((item) => item.count > 0)

    function toggleGroup(date) {
        setExpandedDates((current) => current.includes(date)
            ? current.filter((item) => item !== date)
            : [...current, date])
    }

    return (
        <>
            <section className="dashboard-empty-card dashboard-overview-data-card" aria-labelledby="dashboard-tasks-heading">
                <div className="dashboard-empty-top">
                    <span className="dashboard-tile-icon"><ListTodo size={23} aria-hidden="true" /></span>
                    <span className="dashboard-data-tag">ACROSS YOUR TEAMS</span>
                </div>
                <h3 id="dashboard-tasks-heading">My tasks</h3>
                {state.loading ? <p role="status">Loading assigned work…</p>
                    : state.error ? <p role="alert">{state.error}</p>
                        : <>
                            <div className="dashboard-data-value">{unfinished.length} <span>remaining</span></div>
                            <p>{state.items.length === 0
                                ? 'No tasks assigned yet. Your next assignment will appear here.'
                                : `${completed} completed · ${state.items.length} assigned in total`}</p>
                        </>}
                <button type="button" className="dashboard-card-link" onClick={onOpenTasks}>
                    Open my tasks <ArrowRight size={16} aria-hidden="true" />
                </button>
            </section>

            <section className="dashboard-empty-card dashboard-overview-data-card dashboard-deadline-card"
                aria-labelledby="dashboard-deadlines-heading">
                <div className="dashboard-empty-top">
                    <span className="dashboard-tile-icon"><CalendarDays size={23} aria-hidden="true" /></span>
                    <span className="dashboard-data-tag">ASSIGNED TO YOU</span>
                </div>
                <h3 id="dashboard-deadlines-heading">Your deadlines</h3>
                {state.loading ? <p role="status">Checking your deadlines…</p>
                    : state.error ? <p role="alert">Deadline data is unavailable.</p>
                        : datedCount === 0 ? <>
                            <CheckCircle2 size={22} className="dashboard-deadline-clear" aria-hidden="true" />
                            <p>No unfinished assigned tasks have due dates.</p>
                        </> : <>
                            <div className="dashboard-deadline-summary" aria-label="Grouped deadline reminders">
                                {reminders.map(({ kind, count, label }) => (
                                    <span key={kind} className={`dashboard-deadline-reminder dashboard-deadline-reminder--${kind}`}>
                                        <strong>{count}</strong> {count === 1 ? 'task' : 'tasks'} {label}
                                    </span>
                                ))}
                            </div>
                        </>}
                {datedCount > 0 && !state.loading && !state.error ? (
                    <button type="button" className="dashboard-card-link"
                        onClick={() => setDetailsOpen((open) => !open)}
                        aria-expanded={detailsOpen} aria-controls="dashboard-deadline-details">
                        {detailsOpen ? 'Hide deadline details' : 'View all deadlines'}
                        <ChevronDown size={16} className={detailsOpen ? 'dashboard-chevron--open' : ''} aria-hidden="true" />
                    </button>
                ) : (
                    <button type="button" className="dashboard-card-link" onClick={onOpenTasks}>
                        View assigned tasks <ArrowRight size={16} aria-hidden="true" />
                    </button>
                )}
            </section>

            {detailsOpen && datedCount > 0 && !state.loading && !state.error && (
                <section className="dashboard-deadline-details" id="dashboard-deadline-details"
                    aria-labelledby="dashboard-deadline-details-heading">
                    <header className="dashboard-deadline-details-header">
                        <div>
                            <p className="dashboard-eyebrow">YOUR ASSIGNED WORK</p>
                            <h3 id="dashboard-deadline-details-heading">Deadline details</h3>
                            <p>{datedCount} unfinished {datedCount === 1 ? 'task' : 'tasks'} with due dates · Sorted by date, then priority</p>
                        </div>
                        <button type="button" className="dashboard-deadline-close"
                            onClick={() => setDetailsOpen(false)}>Close details</button>
                    </header>
                    <div className="dashboard-deadline-groups">
                        {groups.map((group) => {
                            const expanded = expandedDates.includes(group.date)
                            const id = `deadline-group-${group.date}`
                            return (
                                <div className="dashboard-deadline-group" key={group.date}>
                                    <button type="button" className="dashboard-deadline-group-toggle"
                                        aria-expanded={expanded} aria-controls={expanded ? id : undefined}
                                        onClick={() => toggleGroup(group.date)}>
                                        <span className="dashboard-deadline-group-title">
                                            <span className={`dashboard-deadline-badge dashboard-deadline-badge--${group.timing.kind}`}>{group.timing.label}</span>
                                            <span className="dashboard-deadline-group-date">{group.date}</span>
                                        </span>
                                        <span className="dashboard-deadline-group-count">
                                            {group.items.length} {group.items.length === 1 ? 'task' : 'tasks'}
                                            <ChevronDown size={16} aria-hidden="true" />
                                        </span>
                                    </button>
                                    {expanded && <ul id={id} className="dashboard-deadline-task-list">
                                        {group.items.map((task) => (
                                            <li key={task.id}>
                                                <span className={`dashboard-deadline-task-priority dashboard-deadline-task-priority--${task.priority}`}>{task.priority}</span>
                                                <div>
                                                    <strong>{task.title}</strong>
                                                    <small>{task.workspace_name} / {task.project_name}</small>
                                                </div>
                                            </li>
                                        ))}
                                    </ul>}
                                </div>
                            )
                        })}
                    </div>
                    <button type="button" className="dashboard-card-link" onClick={onOpenTasks}>
                        Open my tasks <ArrowRight size={16} aria-hidden="true" />
                    </button>
                </section>
            )}
        </>
    )
}
