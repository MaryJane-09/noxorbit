import { useEffect, useState } from 'react'
import { Monitor, Moon, Save, Sun, Accessibility } from 'lucide-react'
import { fetchAppearancePreferences, saveAppearancePreferences } from '../api'
import './AppearanceSettings.css'

const THEMES = [
    { id: 'dark', title: 'Dark', description: 'The original Noxorbit midnight purple.', Icon: Moon },
    { id: 'light', title: 'Light', description: 'Soft ivory, plum, and champagne gold.', Icon: Sun },
    { id: 'system', title: 'System', description: 'Match your device appearance automatically.', Icon: Monitor },
]
const DEFAULT = { theme: 'dark', reduced_motion: false }

export default function AppearanceSettings({ appearance, onSaved }) {
    const [saved, setSaved] = useState(null)
    const [draft, setDraft] = useState(appearance || DEFAULT)
    const [loading, setLoading] = useState(true)
    const [saving, setSaving] = useState(false)
    const [error, setError] = useState('')
    const [success, setSuccess] = useState('')

    useEffect(() => {
        const controller = new AbortController()
        fetchAppearancePreferences({ signal: controller.signal }).then((value) => {
            if (controller.signal.aborted) return
            setSaved(value)
            setDraft(value)
            onSaved(value)
            setError('')
        }).catch((err) => {
            if (!controller.signal.aborted) setError(err.message || 'Unable to load appearance preferences.')
        }).finally(() => { if (!controller.signal.aborted) setLoading(false) })
        return () => controller.abort()
    }, [onSaved])

    const changed = saved && (saved.theme !== draft.theme || saved.reduced_motion !== draft.reduced_motion)

    async function save(event) {
        event.preventDefault()
        if (!changed || saving) return
        setSaving(true)
        setSuccess('')
        setError('')
        try {
            const updated = await saveAppearancePreferences(draft)
            setSaved(updated)
            setDraft(updated)
            onSaved(updated)
            setSuccess('Appearance saved. Your choices will be remembered for this account.')
        } catch (err) {
            setError(err.message || 'Could not save appearance settings.')
        } finally { setSaving(false) }
    }

    return <section className="settings-account appearance-settings" aria-labelledby="appearance-heading">
        <div className="settings-section-title"><Sun size={21} aria-hidden="true" />
            <h2 id="appearance-heading">Appearance</h2></div>
        <p className="settings-help">Choose how Noxorbit looks and moves. Your selection applies across the signed-in workspace.</p>
        {!loading && !saved && <p role="alert" className="dashboard-form-error">{error}</p>}
        {loading ? <p role="status">Loading appearance preferences…</p> : saved && <form onSubmit={save}>
            <fieldset className="appearance-theme-options">
                <legend>Theme</legend>
                <div className="appearance-theme-grid">
                    {THEMES.map(({ id, title, description, Icon }) => <label key={id}
                        className={`appearance-theme-choice${draft.theme === id ? ' appearance-theme-choice--selected' : ''}`}>
                        <input type="radio" name="noxorbit-theme" value={id} checked={draft.theme === id}
                            disabled={saving} onChange={() => { setDraft((p) => ({ ...p, theme: id })); setSuccess('') }} />
                        <Icon size={20} aria-hidden="true" />
                        <strong>{title}</strong>
                        <small>{description}</small>
                    </label>)}
                </div>
            </fieldset>
            <label className="appearance-motion-choice">
                <span className="appearance-motion-icon"><Accessibility size={20} aria-hidden="true" /></span>
                <span><strong>Reduce motion</strong>
                    <small>Limit transitions, decorative movement, and smooth scrolling. Your device’s reduced-motion setting is always respected.</small></span>
                <input type="checkbox" checked={draft.reduced_motion} disabled={saving}
                    onChange={(event) => { setDraft((p) => ({ ...p, reduced_motion: event.target.checked })); setSuccess('') }} />
            </label>
            <p className="appearance-description">Your current theme stays in place until you save. System theme updates automatically when your device switches between light and dark.</p>
            <div className="appearance-save">
                <button type="submit" disabled={!changed || saving}><Save size={17} aria-hidden="true" />
                    {saving ? 'Saving…' : 'Save appearance'}</button>
            </div>
        </form>}
        {saved && error && <p role="alert" className="dashboard-form-error">{error}</p>}
        {success && <p role="status" className="settings-success">{success}</p>}
    </section>
}
