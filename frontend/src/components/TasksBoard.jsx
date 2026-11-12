import { useEffect, useRef, useState } from 'react'
import { ArrowLeft, CalendarDays, Check, LoaderCircle, Pencil, Plus, Trash2, X } from 'lucide-react'
import { changeTaskStatus, createTask, deleteTask, fetchTasks, updateTask } from '../api'
import './TasksBoard.css'

const COLUMNS = [
    { id: 'todo', label: 'To Do' },
    { id: 'in_progress', label: 'In Progress' },
    { id: 'done', label: 'Done' },
]
const PRIORITIES = ['low', 'medium', 'high', 'urgent']

function emptyForm() {
    return { title: '', description: '', priority: 'medium', due_date: '', assignee_id: '' }
}

export default function TasksBoard({ workspace, project, user, onBack }) {
    const workspaceID = workspace.id
    const projectID = project.id
    const [data, setData] = useState({ key: '', items: [], error: '' })
    const [reload, setReload] = useState(0)
    const [editor, setEditor] = useState(null)
    const [form, setForm] = useState(emptyForm)
    const [saving, setSaving] = useState(false)
    const [busyID, setBusyID] = useState(null)
    const [error, setError] = useState('')
    const [formError, setFormError] = useState('')
    const [pendingDelete, setPendingDelete] = useState(null)
    const [deleteError, setDeleteError] = useState('')
    const dialogRef = useRef(null)
    const deleteDialogRef = useRef(null)
    const titleRef = useRef(null)
    const key = `${workspaceID}:${projectID}`
    const ready = data.key === key
    const archived = project.status === 'archived'
    const currentUserID = user?.id || null
    const isOwner = workspace.role === 'owner'

    useEffect(() => {
        const controller = new AbortController()
        fetchTasks(workspaceID, projectID, { signal: controller.signal })
            .then((items) => {
                if (!controller.signal.aborted) setData({ key, items, error: '' })
            })
            .catch((err) => {
                if (!controller.signal.aborted) setData({ key, items: [], error: err.message })
            })
        return () => controller.abort()
    }, [workspaceID, projectID, key, reload])

    useEffect(() => {
        const dialog = deleteDialogRef.current
        if (pendingDelete && dialog && !dialog.open) dialog.showModal()
        if (!pendingDelete && dialog?.open) dialog.close()
    }, [pendingDelete])

    function change(field, value) {
        setForm((current) => ({ ...current, [field]: value }))
    }

    function openEditor(task = null) {
        setEditor(task || { id: null })
        setForm(task ? {
            title: task.title,
            description: task.description,
            priority: task.priority,
            due_date: task.due_date || '',
            assignee_id: task.assignee_id || '',
        } : emptyForm())
        setFormError('')
        dialogRef.current?.showModal()
        titleRef.current?.focus()
    }

    function closeEditor() {
        if (!saving) dialogRef.current?.close()
    }

    async function save(event) {
        event.preventDefault()
        if (saving || archived) return
        const title = form.title.trim()
        if (!title || [...title].length > 200 || [...form.description].length > 5000) {
            setFormError('Title must be 1–200 characters and description at most 5000.')
            return
        }
        setSaving(true)
        setFormError('')
        try {
            const details = {
                title,
                description: form.description,
                priority: form.priority,
                due_date: form.due_date || null,
                assignee_id: form.assignee_id || null,
            }
            const saved = editor.id
                ? await updateTask(workspaceID, projectID, editor.id, details)
                : await createTask(workspaceID, projectID, details)
            setData((current) => current.key !== key ? current : ({
                ...current,
                items: editor.id
                    ? current.items.map((item) => item.id === saved.id ? saved : item)
                    : [saved, ...current.items],
            }))
            dialogRef.current?.close()
        } catch (err) {
            setFormError(err.message || 'Could not save task.')
        } finally {
            setSaving(false)
        }
    }

    async function moveTask(task, status) {
        if (busyID || archived || status === task.status) return
        setBusyID(task.id)
        setError('')
        try {
            const updated = await changeTaskStatus(workspaceID, projectID, task.id, status)
            setData((current) => current.key !== key ? current : ({
                ...current, items: current.items.map((item) => item.id === updated.id ? updated : item),
            }))
        } catch (err) {
            setError(err.message || 'Could not update task status.')
        } finally {
            setBusyID(null)
        }
    }

    function askDelete(task) {
        if (busyID || archived) return
        setDeleteError('')
        setPendingDelete(task)
    }

    function cancelDelete() {
        if (!busyID) setPendingDelete(null)
    }

    async function confirmDelete(event) {
        event.preventDefault()
        const task = pendingDelete
        if (!task || busyID || archived) return
        setBusyID(task.id)
        setDeleteError('')
        try {
            await deleteTask(workspaceID, projectID, task.id)
            setData((current) => current.key !== key ? current : ({
                ...current, items: current.items.filter((item) => item.id !== task.id),
            }))
            setPendingDelete(null)
        } catch (err) {
            setDeleteError(err.message || 'Could not delete task.')
        } finally {
            setBusyID(null)
        }
    }

    const items = ready ? data.items : []
    const canEdit = (task) => !archived && (isOwner || task.created_by === currentUserID)
    const canMove = (task) => !archived && (isOwner || task.created_by === currentUserID || task.assignee_id === currentUserID)

    return (
        <section className="tasks-page">
            <button type="button" className="tasks-back" onClick={onBack}><ArrowLeft size={17} /> All projects</button>
            <header className="tasks-heading">
                <div>
                    <p className="dashboard-eyebrow">PROJECT BOARD · {workspace.name}</p>
                    <h1>{project.name}<span>.</span></h1>
                    <p>{project.description || 'Keep every step of your project in one place.'}</p>
                </div>
                {!archived && ready && !data.error && (
                    <button type="button" className="dashboard-feature-action" onClick={() => openEditor()}>
                        <Plus size={17} /> New task
                    </button>
                )}
            </header>
            {archived && <p className="tasks-notice">This project is archived. Tasks are read-only until it is restored.</p>}
            {error && <p className="dashboard-form-error" role="alert">{error}</p>}
            {!ready && <div className="projects-message" role="status"><LoaderCircle className="dashboard-spin" /> Loading tasks…</div>}
            {ready && data.error && (
                <div className="projects-message" role="alert">
                    <h2>Could not load tasks.</h2><p>{data.error}</p>
                    <button type="button" className="dashboard-feature-action" onClick={() => {
                        setData({ key: '', items: [], error: '' })
                        setReload((n) => n + 1)
                    }}>Retry</button>
                </div>
            )}
            {ready && !data.error && (
                <>
                    <div className="tasks-summary">
                        <strong>{items.length} {items.length === 1 ? 'task' : 'tasks'}</strong>
                        <span>{items.filter((task) => task.status === 'done').length} completed</span>
                        <span>{items.length ? Math.round(items.filter((task) => task.status === 'done').length / items.length * 100) + '% complete' : 'No progress yet'}</span>
                    </div>
                    <div className="tasks-columns">
                        {COLUMNS.map((column) => {
                            const tasks = items.filter((task) => task.status === column.id)
                            return (
                                <section key={column.id} className="tasks-column" aria-label={column.label}>
                                    <div className="tasks-column-heading"><h2>{column.label}</h2><span>{tasks.length}</span></div>
                                    {tasks.length === 0 && <p className="tasks-empty">No tasks here yet.</p>}
                                    {tasks.map((task) => (
                                        <article className="tasks-card" key={task.id}>
                                            <div className="tasks-card-top"><span className={`tasks-priority tasks-priority--${task.priority}`}>{task.priority}</span></div>
                                            <h3>{task.title}</h3>
                                            {task.description && <p className="tasks-description">{task.description}</p>}
                                            <div className="tasks-meta">
                                                {task.due_date && <span><CalendarDays size={14} /> {task.due_date}</span>}
                                                <span>{task.assignee_id ? (task.assignee_id === currentUserID ? 'Assigned to you' : 'Assigned to member') : 'Unassigned'}</span>
                                            </div>
                                            <div className="tasks-actions">
                                                {canMove(task) && (
                                                    <label className="tasks-status-label">
                                                        <span>Status</span>
                                                        <select value={task.status} onChange={(event) => moveTask(task, event.target.value)}
                                                            aria-label={`Change status of ${task.title}`} disabled={busyID !== null}>
                                                            {COLUMNS.map((option) => <option value={option.id} key={option.id}>{option.label}</option>)}
                                                        </select>
                                                    </label>
                                                )}
                                                {canEdit(task) && (
                                                    <>
                                                        <button type="button" disabled={busyID !== null} onClick={() => openEditor(task)} aria-label={`Edit ${task.title}`}><Pencil size={15} /> Edit</button>
                                                        <button type="button" disabled={busyID !== null} onClick={() => askDelete(task)} aria-label={`Delete ${task.title}`}><Trash2 size={15} /> Delete</button>
                                                    </>
                                                )}
                                            </div>
                                        </article>
                                    ))}
                                </section>
                            )
                        })}
                    </div>
                </>
            )}
            <dialog ref={deleteDialogRef} className="dashboard-workspace-dialog tasks-delete-dialog"
                aria-labelledby="tasks-delete-title" aria-describedby="tasks-delete-description"
                onCancel={(event) => { if (busyID) event.preventDefault(); else setPendingDelete(null) }}
                onClose={() => { if (!busyID) setPendingDelete(null) }}>
                <div className="dashboard-dialog-header">
                    <span className="dashboard-dialog-icon tasks-delete-icon"><Trash2 size={22} aria-hidden="true" /></span>
                    <button type="button" className="dashboard-dialog-close" onClick={cancelDelete}
                        disabled={Boolean(busyID)} aria-label="Close delete confirmation"><X size={19} /></button>
                </div>
                <p className="dashboard-eyebrow">CONFIRM YOUR ACTION</p>
                <h2 id="tasks-delete-title">Delete task?</h2>
                <p id="tasks-delete-description" className="dashboard-dialog-description">
                    You are about to permanently delete <strong>{pendingDelete?.title}</strong>.
                    This cannot be undone. A recycle bin is planned but is not available yet.
                </p>
                {deleteError && <p role="alert" className="dashboard-form-error">{deleteError}</p>}
                <form onSubmit={confirmDelete} className="dashboard-dialog-actions">
                    <button type="button" className="dashboard-dialog-cancel"
                        disabled={Boolean(busyID)} onClick={cancelDelete}>Cancel</button>
                    <button type="submit" className="tasks-delete-confirm" disabled={Boolean(busyID)}>
                        {busyID ? <><LoaderCircle size={16} className="dashboard-spin" /> Deleting…</>
                            : <><Trash2 size={16} /> Delete task</>}
                    </button>
                </form>
            </dialog>
            <dialog className="dashboard-workspace-dialog tasks-dialog" ref={dialogRef}
                aria-labelledby="tasks-dialog-title"
                onCancel={(event) => { if (saving) event.preventDefault() }}
                onClose={() => { setEditor(null); setFormError('') }}>
                <div className="dashboard-dialog-header">
                    <span className="dashboard-dialog-icon"><Check size={21} /></span>
                    <button type="button" className="dashboard-dialog-close" onClick={closeEditor} disabled={saving} aria-label="Close task form"><X size={19} /></button>
                </div>
                <p className="dashboard-eyebrow">MAKE PROGRESS VISIBLE</p>
                <h2 id="tasks-dialog-title">{editor?.id ? 'Edit task.' : 'Create a task.'}</h2>
                <form className="dashboard-workspace-form" onSubmit={save}>
                    <label htmlFor="task-title">Task title</label>
                    <input id="task-title" ref={titleRef} value={form.title} onChange={(event) => change('title', event.target.value)}
                        maxLength={200} required disabled={saving} placeholder="e.g. Prepare presentation" />
                    <label htmlFor="task-description">Description (optional)</label>
                    <textarea id="task-description" rows={3} maxLength={5000} disabled={saving} value={form.description}
                        onChange={(event) => change('description', event.target.value)} placeholder="What needs to be done?" />
                    <label htmlFor="task-priority">Priority</label>
                    <select id="task-priority" value={form.priority} onChange={(event) => change('priority', event.target.value)} disabled={saving}>
                        {PRIORITIES.map((value) => <option key={value} value={value}>{value.charAt(0).toUpperCase() + value.slice(1)}</option>)}
                    </select>
                    <label htmlFor="task-due">Due date (optional)</label>
                    <input id="task-due" type="date" value={form.due_date} onChange={(event) => change('due_date', event.target.value)} disabled={saving} />
                    <label htmlFor="task-assignee">Assignee</label>
                    <select id="task-assignee" value={form.assignee_id}
                        disabled={saving || (Boolean(form.assignee_id) && form.assignee_id !== currentUserID)}
                        onChange={(event) => change('assignee_id', event.target.value)}>
                        <option value="">Unassigned</option>
                        {currentUserID && <option value={currentUserID}>Assign to myself</option>}
                        {form.assignee_id && form.assignee_id !== currentUserID && <option value={form.assignee_id}>Assigned to another member</option>}
                    </select>
                    <p className="dashboard-form-hint">Assigning other teammates will be available when the workspace member directory is connected.</p>
                    {formError && <p role="alert" className="dashboard-form-error">{formError}</p>}
                    <div className="dashboard-dialog-actions">
                        <button type="button" className="dashboard-dialog-cancel" disabled={saving} onClick={closeEditor}>Cancel</button>
                        <button type="submit" className="dashboard-feature-action dashboard-dialog-submit" disabled={saving || !form.title.trim()}>
                            {saving ? <><LoaderCircle className="dashboard-spin" size={16} /> Saving…</> : <><Check size={16} /> Save task</>}
                        </button>
                    </div>
                </form>
            </dialog>
        </section>
    )
}
