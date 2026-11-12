import { useEffect, useRef, useState } from 'react'
import { ArchiveRestore, LoaderCircle, RefreshCw, Trash2, X } from 'lucide-react'
import { deleteRecycledTaskForever, fetchRecycledTasks, restoreRecycledTask } from '../api'
import './RecycleBinPage.css'

// Only tasks are recoverable in this phase; project archiving remains separate.
export default function RecycleBinPage() {
    const [state, setState] = useState({ items: [], loading: true, error: '' })
    const [version, setVersion] = useState(0)
    const [busyID, setBusyID] = useState(null)
    const [actionError, setActionError] = useState('')
    const [notice, setNotice] = useState('')
    const [pendingDelete, setPendingDelete] = useState(null)
    const confirmDialogRef = useRef(null)

    useEffect(() => {
        const controller = new AbortController()
        fetchRecycledTasks({ signal: controller.signal })
            .then((items) => {
                if (!controller.signal.aborted) setState({ items, loading: false, error: '' })
            })
            .catch((error) => {
                if (!controller.signal.aborted) setState({ items: [], loading: false, error: error.message })
            })
        return () => controller.abort()
    }, [version])

    useEffect(() => {
        const dialog = confirmDialogRef.current
        if (pendingDelete && dialog && !dialog.open) dialog.showModal()
        if (!pendingDelete && dialog?.open) dialog.close()
    }, [pendingDelete])

    function refresh() {
        setActionError('')
        setState((current) => ({ ...current, loading: true, error: '' }))
        setVersion((value) => value + 1)
    }

    async function restore(item) {
        if (busyID) return
        setBusyID(item.id)
        setActionError('')
        setNotice('')
        try {
            await restoreRecycledTask(item.id)
            setState((current) => ({
                ...current,
                items: current.items.filter((entry) => entry.id !== item.id),
            }))
            setNotice(`Restored "${item.title}" to its project board.`)
        } catch (error) {
            setActionError(error.message || 'Could not restore task.')
        } finally {
            setBusyID(null)
        }
    }

    async function deleteForever(event) {
        event.preventDefault()
        if (!pendingDelete || busyID) return
        const task = pendingDelete
        setBusyID(task.id)
        setActionError('')
        setNotice('')
        try {
            await deleteRecycledTaskForever(task.id)
            setState((current) => ({
                ...current,
                items: current.items.filter((entry) => entry.id !== task.id),
            }))
            setPendingDelete(null)
            setNotice(`Permanently deleted "${task.title}". This action cannot be undone.`)
        } catch (error) {
            setActionError(error.message || 'Could not permanently delete task.')
        } finally {
            setBusyID(null)
        }
    }

    return <section className="recycle-page" aria-labelledby="recycle-heading">
        <header className="recycle-heading">
            <div>
                <p className="dashboard-eyebrow">RECOVER YOUR WORK</p>
                <h1 id="recycle-heading">Recycle bin<span>.</span></h1>
                <p>Restore deleted tasks to their original projects.</p>
            </div>
            <button type="button" className="recycle-refresh" onClick={refresh} disabled={state.loading || Boolean(busyID)}>
                <RefreshCw size={17} aria-hidden="true" /> Refresh
            </button>
        </header>
        {actionError && <p role="alert" className="dashboard-form-error">{actionError}</p>}
        {notice && <p role="status" className="recycle-notice">{notice}</p>}
        {state.loading ? <div className="recycle-message" role="status"><LoaderCircle size={19} className="dashboard-spin" /> Loading deleted tasks…</div>
            : state.error ? <div className="recycle-message" role="alert">
                <p>{state.error}</p><button type="button" onClick={refresh}>Try again</button>
            </div> : state.items.length === 0 ? <div className="recycle-message">
                <Trash2 size={30} aria-hidden="true" />
                <h2>Your recycle bin is empty.</h2>
                <p>Deleted tasks you can restore will appear here.</p>
            </div> : <div className="recycle-items">
                {state.items.map((task) => <article className="recycle-item" key={task.id}>
                    <div className="recycle-item-info">
                        <span>{task.workspace_name} / {task.project_name}</span>
                        <h2>{task.title}</h2>
                        <p>Deleted {new Date(task.deleted_at).toLocaleString()}</p>
                    </div>
                    <div className="recycle-item-actions">
                        <button type="button" className="recycle-restore" onClick={() => restore(task)} disabled={Boolean(busyID)}>
                            {busyID === task.id ? <LoaderCircle size={17} className="dashboard-spin" aria-hidden="true" />
                                : <ArchiveRestore size={17} aria-hidden="true" />}
                            Restore task
                        </button>
                        <button type="button" className="recycle-delete-forever"
                            onClick={() => { setActionError(''); setPendingDelete(task) }}
                            disabled={Boolean(busyID)}>
                            <Trash2 size={16} aria-hidden="true" /> Delete forever
                        </button>
                    </div>
                </article>)}
            </div>}
        <dialog ref={confirmDialogRef} className="dashboard-workspace-dialog recycle-delete-dialog"
            aria-labelledby="recycle-delete-title" aria-describedby="recycle-delete-description"
            onCancel={(event) => { if (busyID) event.preventDefault(); else setPendingDelete(null) }}
            onClose={() => { if (!busyID) setPendingDelete(null) }}>
            <div className="dashboard-dialog-header">
                <span className="dashboard-dialog-icon recycle-delete-icon"><Trash2 size={22} aria-hidden="true" /></span>
                <button type="button" className="dashboard-dialog-close"
                    disabled={Boolean(busyID)} onClick={() => setPendingDelete(null)}
                    aria-label="Close permanent deletion confirmation"><X size={19} aria-hidden="true" /></button>
            </div>
            <p className="dashboard-eyebrow">PERMANENT DELETION</p>
            <h2 id="recycle-delete-title">Delete forever?</h2>
            <p id="recycle-delete-description" className="dashboard-dialog-description">
                Permanently delete <strong>{pendingDelete?.title}</strong> from
                <strong> {pendingDelete?.workspace_name} / {pendingDelete?.project_name}</strong>?
                This removes the task permanently and it cannot be restored.
            </p>
            {actionError && <p role="alert" className="dashboard-form-error">{actionError}</p>}
            <form className="dashboard-dialog-actions" onSubmit={deleteForever}>
                <button type="button" className="dashboard-dialog-cancel" disabled={Boolean(busyID)}
                    onClick={() => setPendingDelete(null)}>Cancel</button>
                <button type="submit" className="recycle-confirm-delete" disabled={Boolean(busyID)}>
                    {busyID ? <><LoaderCircle size={16} className="dashboard-spin" /> Deleting…</>
                        : <><Trash2 size={16} aria-hidden="true" /> Delete forever</>}
                </button>
            </form>
        </dialog>
    </section>
}
