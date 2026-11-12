import { useEffect, useState } from 'react'
import { ArrowRight, CalendarDays, CheckCircle2, ListTodo } from 'lucide-react'
import { fetchMyTasks } from '../api'

// The overview reuses the authenticated My Tasks endpoint rather than
// gathering every project and task with a cascade of client requests.
export default function OverviewTaskCards({ onOpenTasks }) {
    const [state, setState] = useState({ items: [], loading: true, error: '' })

    useEffect(() => {
        const controller = new AbortController()
        fetchMyTasks({ signal: controller.signal })
            .then((items) => {
                if (!controller.signal.aborted) setState({ items, loading: false, error: '' })
            })
            .catch((error) => {
                if (!controller.signal.aborted) setState({ items: [], loading: false, error: error.message || 'Could not load your tasks.' })
            })
        return () => controller.abort()
    }, [])

    const unfinished = state.items.filter((task) => task.status !== 'done')
    const completed = state.items.length - unfinished.length
    const dated = unfinished.filter((task) => task.due_date)
        .sort((a, b) => a.due_date.localeCompare(b.due_date) || a.id.localeCompare(b.id))
    const nextDue = dated[0]

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

            <section className="dashboard-empty-card dashboard-overview-data-card" aria-labelledby="dashboard-deadlines-heading">
                <div className="dashboard-empty-top">
                    <span className="dashboard-tile-icon"><CalendarDays size={23} aria-hidden="true" /></span>
                    <span className="dashboard-data-tag">ASSIGNED TO YOU</span>
                </div>
                <h3 id="dashboard-deadlines-heading">Your upcoming deadlines</h3>
                {state.loading ? <p role="status">Checking your deadlines…</p>
                    : state.error ? <p role="alert">Deadline data is unavailable.</p>
                        : nextDue ? <>
                            <div className="dashboard-deadline-date">{nextDue.due_date}</div>
                            <p className="dashboard-deadline-label">{nextDue.title}</p>
                            <p className="dashboard-deadline-context">{nextDue.workspace_name} / {nextDue.project_name}</p>
                        </> : <>
                            <CheckCircle2 size={22} className="dashboard-deadline-clear" aria-hidden="true" />
                            <p>No unfinished assigned tasks have due dates.</p>
                        </>}
                <button type="button" className="dashboard-card-link" onClick={onOpenTasks}>
                    View assigned work <ArrowRight size={16} aria-hidden="true" />
                </button>
            </section>
        </>
    )
}
