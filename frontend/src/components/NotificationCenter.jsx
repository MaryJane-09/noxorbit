import { useCallback, useEffect, useState } from 'react'
import { Bell, CheckCheck, RefreshCw } from 'lucide-react'
import { fetchNotifications, markAllNotificationsRead, markNotificationRead } from '../api'
import './NotificationCenter.css'

export default function NotificationCenter() {
    const [items, setItems] = useState([])
    const [open, setOpen] = useState(false)
    const [error, setError] = useState('')
    const [loading, setLoading] = useState(true)
    const [busy, setBusy] = useState(false)

    const refresh = useCallback(async (signal) => {
        try {
            const result = await fetchNotifications({ signal })
            if (signal?.aborted) return
            setItems(result)
            setError('')
        } catch (err) {
            if (signal?.aborted) return
            setError(err.message || 'Could not load notifications.')
        } finally {
            if (!signal?.aborted) setLoading(false)
        }
    }, [])

    useEffect(() => {
        const controller = new AbortController()
        // Schedule the first fetch outside the synchronous effect body.
        // The abort signal also prevents updates after unmount.
        const initialTimer = window.setTimeout(() => {
            void refresh(controller.signal)
        }, 0)
        const pollTimer = window.setInterval(() => {
            if (!document.hidden) void refresh(controller.signal)
        }, 60_000)
        const onPreferencesSaved = () => { void refresh(controller.signal) }
        window.addEventListener('noxorbit:notification-preferences-saved', onPreferencesSaved)
        return () => {
            controller.abort()
            window.clearTimeout(initialTimer)
            window.clearInterval(pollTimer)
            window.removeEventListener('noxorbit:notification-preferences-saved', onPreferencesSaved)
        }
    }, [refresh])

    const unread = items.filter((item) => !item.read_at).length
    async function markOne(id) {
        if (busy) return
        setBusy(true)
        try {
            await markNotificationRead(id)
            setItems((current) => current.map((item) => item.id === id ? { ...item, read_at: new Date().toISOString() } : item))
        } catch (err) {
            setError(err.message || 'Could not update notification.')
        } finally { setBusy(false) }
    }
    async function markAll() {
        if (busy) return
        setBusy(true)
        try {
            await markAllNotificationsRead()
            setItems((current) => current.map((item) => ({ ...item, read_at: item.read_at || new Date().toISOString() })))
        } catch (err) {
            setError(err.message || 'Could not update notifications.')
        } finally { setBusy(false) }
    }

    return <div className="notification-center">
        <button type="button" className="notification-bell" aria-label={`Notifications, ${unread} unread`}
            aria-expanded={open} aria-controls="notification-panel" title="Notifications"
            onClick={() => { setOpen((value) => !value); refresh() }}>
            <Bell size={19} aria-hidden="true" />
            {unread > 0 && <span className="notification-unread-count">{unread > 99 ? '99+' : unread}</span>}
        </button>
        {open && <div id="notification-panel" className="notification-panel" aria-label="Your notifications">
            <div className="notification-panel-header">
                <div><strong>Notifications</strong><small>Updates for your account</small></div>
                <button type="button" onClick={() => refresh()} title="Refresh notifications" aria-label="Refresh notifications"><RefreshCw size={17} /></button>
            </div>
            {error && <p className="notification-error" role="alert">{error}</p>}
            <div className="notification-panel-actions">
                <span>{unread} unread</span>
                {unread > 0 && <button type="button" onClick={markAll} disabled={busy}>
                    <CheckCheck size={15} /> Mark all read</button>}
            </div>
            <div className="notification-items">
                {loading ? <p className="notification-empty">Loading notifications…</p>
                    : items.length === 0 ? <p className="notification-empty">You're all caught up. New task assignments and deadline reminders will appear here.</p>
                        : items.map((item) => <article key={item.id} className={`notification-item${item.read_at ? '' : ' notification-item--unread'}`}>
                            <div><strong>{item.title}</strong><p>{item.body}</p>
                                <small>{new Date(item.created_at).toLocaleString()}</small></div>
                            {!item.read_at && <button type="button" disabled={busy}
                                onClick={() => markOne(item.id)}>Mark read</button>}
                        </article>)}
            </div>
            <p className="notification-panel-footnote">Deadline reminders update while you're using Noxorbit. Email delivery isn't available yet.</p>
        </div>}
    </div>
}
