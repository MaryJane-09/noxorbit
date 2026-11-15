import { useEffect, useState } from 'react'
import { BellRing, Save } from 'lucide-react'
import { fetchNotificationPreferences, saveNotificationPreferences } from '../api'
import './NotificationSettings.css'

const CHOICES = [
    { key: 'task_assigned', title: 'Task assigned to me', description: 'Notify me when another person assigns me a task.' },
    { key: 'due_tomorrow', title: 'Task due tomorrow', description: 'Remind me one calendar day before an assigned task is due.' },
    { key: 'due_today', title: 'Task due today', description: 'Remind me on an assigned task’s due date.' },
    { key: 'overdue', title: 'Overdue tasks', description: 'Remind me about unfinished assigned tasks after their due date.' },
]

export default function NotificationSettings() {
    const [settings, setSettings] = useState(null)
    const [draft, setDraft] = useState(null)
    const [error, setError] = useState('')
    const [message, setMessage] = useState('')
    const [saving, setSaving] = useState(false)
    const [loading, setLoading] = useState(true)
    useEffect(() => {
        const controller = new AbortController()
        fetchNotificationPreferences({ signal: controller.signal })
            .then((data) => {
                if (controller.signal.aborted) return
                setSettings(data)
                setDraft(data)
                setError('')
            })
            .catch((err) => { if (!controller.signal.aborted) setError(err.message || 'Could not load preferences.') })
            .finally(() => { if (!controller.signal.aborted) setLoading(false) })
        return () => controller.abort()
    }, [])
    const dirty = settings && draft && CHOICES.some((choice) => settings[choice.key] !== draft[choice.key])
    async function save(event) {
        event.preventDefault()
        if (!dirty || saving) return
        setSaving(true)
        setError('')
        setMessage('')
        try {
            const updated = await saveNotificationPreferences(draft)
            setSettings(updated)
            setDraft(updated)
            // Refresh the header bell immediately after a successful save.
            window.dispatchEvent(new Event('noxorbit:notification-preferences-saved'))
            setMessage('Notification preferences saved.')
        } catch (err) {
            setError(err.message || 'Could not save notification preferences.')
        } finally { setSaving(false) }
    }

    return <section className="notification-settings settings-account" aria-labelledby="notification-settings-heading">
        <div className="settings-section-title"><BellRing size={21} aria-hidden="true" />
            <h2 id="notification-settings-heading">Notifications</h2></div>
        <p className="settings-help">Choose which in-app notifications you receive. Select Save preferences after each change. Re-enabling a deadline reminder can make today's existing reminder unread again if the task is still due.</p>
        <p className="notification-settings-note">Notifications appear in the bell at the top of your dashboard. Deadline reminders are checked while you're using Noxorbit; background and email delivery are not active yet. Deadline reminders follow your time zone under Settings → Preferences.</p>
        {loading ? <p role="status">Loading preferences…</p>
            : draft && <form onSubmit={save}>
                {CHOICES.map((choice) => <label key={choice.key} className="notification-settings-row">
                    <span><strong>{choice.title}</strong><small>{choice.description}</small></span>
                    <input type="checkbox" checked={draft[choice.key]}
                        onChange={(event) => {
                            setMessage('')
                            setDraft((current) => ({ ...current, [choice.key]: event.target.checked }))
                        }} disabled={saving} />
                </label>)}
                <div className="notification-settings-actions">
                    <button type="submit" disabled={!dirty || saving}><Save size={16} />
                        {saving ? 'Saving…' : 'Save preferences'}</button>
                </div>
            </form>}
        {error && <p className="dashboard-form-error" role="alert">{error}</p>}
        {message && <p className="settings-success" role="status">{message}</p>}
    </section>
}
