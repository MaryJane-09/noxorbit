import { useEffect, useState } from 'react'
import { CalendarDays, Save } from 'lucide-react'
import { fetchUserPreferences, saveUserPreferences } from '../api'
import { DEFAULT_PREFERENCES, calendarToday, formatCalendarDate } from './deadlines'
import './UserPreferences.css'

const ZONES = [
    ['Africa/Lagos', 'Lagos (WAT)'],
    ['UTC', 'UTC'],
    ['Europe/London', 'London'],
    ['Europe/Paris', 'Paris'],
    ['America/New_York', 'New York'],
    ['America/Los_Angeles', 'Los Angeles'],
    ['Asia/Dubai', 'Dubai'],
    ['Asia/Kolkata', 'India'],
    ['Asia/Tokyo', 'Tokyo'],
]
const FORMATS = [
    ['YYYY-MM-DD', '2026-11-15'],
    ['DD/MM/YYYY', '15/11/2026'],
    ['MM/DD/YYYY', '11/15/2026'],
]

export default function UserPreferences({ preferences, onSaved }) {
    const [draft, setDraft] = useState(preferences || DEFAULT_PREFERENCES)
    const [saved, setSaved] = useState(null)
    const [loaded, setLoaded] = useState(false)
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState('')
    const [notice, setNotice] = useState('')

    useEffect(() => {
        const controller = new AbortController()
        fetchUserPreferences({ signal: controller.signal }).then((value) => {
            if (!controller.signal.aborted) {
                setDraft(value)
                setSaved(value)
                onSaved(value)
                setLoaded(true)
                setError('')
            }
        }).catch((err) => {
            if (!controller.signal.aborted) { setError(err.message || 'Could not load preferences.'); setLoaded(false) }
        })
        return () => controller.abort()
    }, [onSaved])

    const dirty = loaded && saved && (draft.time_zone !== saved.time_zone || draft.date_format !== saved.date_format)
    const today = calendarToday(new Date(), draft.time_zone)
    async function save(event) {
        event.preventDefault()
        if (!dirty || saving) return
        setSaving(true)
        setError('')
        setNotice('')
        try {
            const saved = await saveUserPreferences(draft)
            setDraft(saved)
            setSaved(saved)
            onSaved(saved)
            setNotice('Preferences saved. Deadline labels and dates now use your choices.')
        } catch (err) {
            setError(err.message || 'Could not save preferences.')
        } finally { setSaving(false) }
    }

    return <section className="settings-account user-preferences" aria-labelledby="user-preferences-heading">
        <div className="settings-section-title"><CalendarDays size={21} aria-hidden="true" />
            <h2 id="user-preferences-heading">Preferences</h2></div>
        <p className="settings-help">Choose the calendar time zone Noxorbit uses for your task reminders, and how dates are displayed.</p>
        <p className="user-preferences-note">Task deadlines are saved as calendar dates, not times. Changing your time zone changes what counts as today or tomorrow for your account; it does not move the task's due date.</p>
        {!loaded ? (!error && <p role="status">Loading preferences…</p>) : <form onSubmit={save}>
            <label htmlFor="user-preference-zone">Time zone</label>
            <select id="user-preference-zone" value={draft.time_zone} disabled={saving}
                onChange={(event) => setDraft((current) => ({ ...current, time_zone: event.target.value }))}>
                {ZONES.map(([id, label]) => <option key={id} value={id}>{label} · {id}</option>)}
            </select>
            <label htmlFor="user-preference-format">Date format</label>
            <select id="user-preference-format" value={draft.date_format} disabled={saving}
                onChange={(event) => setDraft((current) => ({ ...current, date_format: event.target.value }))}>
                {FORMATS.map(([id, label]) => <option key={id} value={id}>{label} ({id})</option>)}
            </select>
            <div className="user-preferences-preview">
                <strong>Preview</strong>
                <span>Today: {formatCalendarDate(today, draft.date_format)}</span>
                <small>Your saved time zone will also control deadline reminder categories.</small>
            </div>
            <div className="user-preferences-actions"><button type="submit" disabled={!dirty || saving}>
                <Save size={16} /> {saving ? 'Saving…' : 'Save preferences'}
            </button></div>
        </form>}
        {error && <p role="alert" className="dashboard-form-error">{error}</p>}
        {notice && <p role="status" className="settings-success">{notice}</p>}
    </section>
}
