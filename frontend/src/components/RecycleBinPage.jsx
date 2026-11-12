import { useEffect, useState } from 'react'
import { ArchiveRestore, LoaderCircle, RefreshCw, Trash2 } from 'lucide-react'
import { fetchRecycledTasks, restoreRecycledTask } from '../api'
import './RecycleBinPage.css'

// Only tasks are recoverable in this phase; project archiving remains separate.
export default function RecycleBinPage() {
    const [state, setState] = useState({ items: [], loading: true, error: '' })
    const [version, setVersion] = useState(0)
    const [busyID, setBusyID] = useState(null)
    const [actionError, setActionError] = useState('')
    const [notice, setNotice] = useState('')

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

    return <section className="recycle-page" aria-labelledby="recycle-heading">
        <header className="recycle-heading">
            <div>
                <p className="dashboard-eyebrow">RECOVER YOUR WORK</p>
                <h1 id="recycle-heading">Recycle bin<span>.</span></h1>
                <p>Restore deleted tasks to their original projects. Only tasks you can recover appear here.</p>
            </div>
            <button type="button" className="recycle-refresh" onClick={refresh} disabled={state.loading || Boolean(busyID)}>
                <RefreshCw size={17} aria-hidden="true" /> Refresh
            </button>
        </header>
        <p className="recycle-scope">Task recovery is available now. Project archiving remains available under Projects; workspace deletion is not enabled.</p>
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
                    <button type="button" onClick={() => restore(task)} disabled={Boolean(busyID)}>
                        {busyID === task.id ? <LoaderCircle size={17} className="dashboard-spin" aria-hidden="true" />
                            : <ArchiveRestore size={17} aria-hidden="true" />}
                        Restore task
                    </button>
                </article>)}
            </div>}
    </section>
}
