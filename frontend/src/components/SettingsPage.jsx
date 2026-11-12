import { useEffect, useState } from 'react'
import { AlertTriangle, LoaderCircle, RefreshCw, ShieldAlert, Trash2, UserRound, KeyRound } from 'lucide-react'
import { changeAccountPassword, deleteMyAccount, fetchAccountDeletionImpact, saveProfileName } from '../api'
import './SettingsPage.css'

export default function SettingsPage({ user, onAccountDeleted, onProfileUpdated, onPasswordChanged }) {
    const [tab, setTab] = useState('profile')
    const [displayName, setDisplayName] = useState(user.name || '')
    const [savingName, setSavingName] = useState(false)
    const [profileError, setProfileError] = useState('')
    const [profileMessage, setProfileMessage] = useState('')
    const [oldPassword, setOldPassword] = useState('')
    const [newPassword, setNewPassword] = useState('')
    const [repeatedPassword, setRepeatedPassword] = useState('')
    const [changingPassword, setChangingPassword] = useState(false)
    const [passwordError, setPasswordError] = useState('')
    const [state, setState] = useState({ impact: null, loading: true, error: '' })
    const [version, setVersion] = useState(0)
    const [password, setPassword] = useState('')
    const [phrase, setPhrase] = useState('')
    const [consent, setConsent] = useState(false)
    const [busy, setBusy] = useState(false)
    const [submitError, setSubmitError] = useState('')
    const [open, setOpen] = useState(false)

    useEffect(() => {
        const controller = new AbortController()
        fetchAccountDeletionImpact({ signal: controller.signal })
            .then((impact) => {
                if (!controller.signal.aborted) setState({ impact, loading: false, error: '' })
            })
            .catch((error) => {
                if (!controller.signal.aborted) setState({ impact: null, loading: false, error: error.message })
            })
        return () => controller.abort()
    }, [version])

    const sole = state.impact?.sole_owned || []
    const shared = state.impact?.shared_owned || []
    const blocked = shared.length > 0
    const canDelete = !busy && !state.loading && !state.error && !blocked
        && password.length > 0 && phrase === 'DELETE MY ACCOUNT'
        && (sole.length === 0 || consent)

    async function saveName(event) {
        event.preventDefault()
        if (savingName) return
        const name = displayName.trim().replace(/\s+/g, ' ')
        if (name.length < 1 || [...name].length > 100) {
            setProfileError('Name must be between 1 and 100 characters.')
            return
        }
        setSavingName(true)
        setProfileError('')
        setProfileMessage('')
        try {
            const profile = await saveProfileName(name)
            onProfileUpdated(profile)
            setDisplayName(profile.name)
            setProfileMessage('Your name has been updated.')
        } catch (error) {
            setProfileError(error.message || 'Could not save your name.')
        } finally {
            setSavingName(false)
        }
    }

    async function updatePassword(event) {
        event.preventDefault()
        if (changingPassword) return
        if (newPassword !== repeatedPassword) {
            setPasswordError('New passwords do not match.')
            return
        }
        if (newPassword.length < 8 || newPassword.length > 72) {
            setPasswordError('New password must be between 8 and 72 characters.')
            return
        }
        setChangingPassword(true)
        setPasswordError('')
        try {
            await changeAccountPassword(oldPassword, newPassword)
            setOldPassword('')
            setNewPassword('')
            setRepeatedPassword('')
            onPasswordChanged()
        } catch (error) {
            setPasswordError(error.message || 'Could not change password.')
        } finally {
            setChangingPassword(false)
        }
    }

    async function confirmDeletion(event) {
        event.preventDefault()
        if (!canDelete) return
        setSubmitError('')
        setBusy(true)
        try {
            await deleteMyAccount({ password, deleteSoleOwnedWorkspaces: sole.length > 0 && consent })
            onAccountDeleted()
        } catch (error) {
            setSubmitError(error.message || 'Account deletion failed.')
        } finally {
            setBusy(false)
        }
    }

    function refresh() {
        setState({ impact: null, loading: true, error: '' })
        setVersion((value) => value + 1)
    }

    return <section className="settings-page" aria-labelledby="settings-heading">
        <header className="settings-heading">
            <p className="dashboard-eyebrow">PERSONAL SETTINGS</p>
            <h1 id="settings-heading">Settings<span>.</span></h1>
            <p>Manage your account and review important account actions.</p>
        </header>
        <nav className="settings-tabs" aria-label="Settings sections">
            <button type="button" className={tab === 'profile' ? 'settings-tab--active' : ''}
                aria-current={tab === 'profile' ? 'page' : undefined} onClick={() => setTab('profile')}>
                <UserRound size={18} aria-hidden="true" /> My profile
            </button>
            <button type="button" className={tab === 'security' ? 'settings-tab--active' : ''}
                aria-current={tab === 'security' ? 'page' : undefined} onClick={() => setTab('security')}>
                <KeyRound size={18} aria-hidden="true" /> Security
            </button>
        </nav>
        {tab === 'profile' && <section className="settings-account" aria-labelledby="settings-account-heading">
            <div className="settings-section-title">
                <UserRound size={21} aria-hidden="true" />
                <h2 id="settings-account-heading">My profile</h2>
            </div>
            <p className="settings-help">Your name is visible to teammates across Noxorbit.</p>
            <form onSubmit={saveName} className="settings-edit-form">
                <label htmlFor="settings-display-name">Full name</label>
                <input id="settings-display-name" value={displayName} autoComplete="name" maxLength={100}
                    onChange={(event) => setDisplayName(event.target.value)} />
                <label htmlFor="settings-email-display">Email address</label>
                <input id="settings-email-display" value={user.email} readOnly aria-describedby="settings-email-note" />
                <p id="settings-email-note" className="settings-help">Email changes will require verification and are not enabled yet.</p>
                {profileError && <p role="alert" className="settings-deletion-error">{profileError}</p>}
                {profileMessage && <p role="status" className="settings-success">{profileMessage}</p>}
                <div className="settings-edit-actions">
                    <button type="submit" disabled={savingName || displayName.trim() === user.name}>
                        {savingName ? 'Saving…' : 'Save changes'}
                    </button>
                </div>
            </form>
        </section>}
        {tab === 'security' && <>
            <section className="settings-account" aria-labelledby="settings-security-heading">
                <div className="settings-section-title">
                    <KeyRound size={21} aria-hidden="true" />
                    <h2 id="settings-security-heading">Change password</h2>
                </div>
                <p className="settings-help">After you change your password, Noxorbit signs you out of all devices. Sign in again using your new password.</p>
                <form onSubmit={updatePassword} className="settings-edit-form">
                    <label htmlFor="settings-current-password">Current password</label>
                    <input id="settings-current-password" type="password" autoComplete="current-password"
                        value={oldPassword} onChange={(event) => setOldPassword(event.target.value)} required />
                    <label htmlFor="settings-new-password">New password</label>
                    <input id="settings-new-password" type="password" autoComplete="new-password"
                        value={newPassword} onChange={(event) => setNewPassword(event.target.value)} required minLength={8} maxLength={72} />
                    <label htmlFor="settings-repeat-password">Confirm new password</label>
                    <input id="settings-repeat-password" type="password" autoComplete="new-password"
                        value={repeatedPassword} onChange={(event) => setRepeatedPassword(event.target.value)} required />
                    {passwordError && <p role="alert" className="settings-deletion-error">{passwordError}</p>}
                    <div className="settings-edit-actions">
                        <button type="submit" disabled={changingPassword || !oldPassword || !newPassword || !repeatedPassword}>
                            {changingPassword ? 'Changing password…' : 'Change password'}
                        </button>
                    </div>
                </form>
            </section>
        <section className="settings-danger" aria-labelledby="settings-danger-heading">
            <div className="settings-section-title">
                <ShieldAlert size={21} aria-hidden="true" />
                <h2 id="settings-danger-heading">Delete account</h2>
            </div>
            <p>Deleting your Noxorbit account is permanent. You will be signed out on all devices. Your membership in other teams will end, but their projects and tasks will remain.</p>
            {!open ? <button type="button" className="settings-open-danger" onClick={() => setOpen(true)}>
                <Trash2 size={17} aria-hidden="true" /> Review account deletion
            </button> : <div className="settings-deletion-review">
                <div className="settings-review-heading">
                    <h3>Review what will happen</h3>
                    <button type="button" onClick={refresh} disabled={state.loading || busy}>
                        <RefreshCw size={16} aria-hidden="true" /> Refresh
                    </button>
                </div>
                {state.loading ? <p role="status"><LoaderCircle size={17} className="dashboard-spin" aria-hidden="true" /> Checking workspace ownership…</p>
                    : state.error ? <p role="alert" className="settings-deletion-error">{state.error}</p>
                    : <>
                        {blocked && <div className="settings-owner-warning" role="alert">
                            <AlertTriangle size={20} aria-hidden="true" />
                            <div><strong>Account deletion is blocked.</strong>
                                <p>These workspaces have other members and still need an owner. Transfer ownership before deleting your account.</p>
                                <ul>{shared.map((item) => <li key={item.id}>{item.name} ({item.member_count} members)</li>)}</ul>
                            </div>
                        </div>}
                        {sole.length > 0 && <div className="settings-solo-warning">
                            <strong>{sole.length} workspace{sole.length === 1 ? '' : 's'} where you are the only member</strong>
                            <p>If you delete your account, these workspaces and their projects, tasks, invitations, and activity history will also be permanently deleted:</p>
                            <ul>{sole.map((item) => <li key={item.id}>{item.name}</li>)}</ul>
                        </div>}
                        {!blocked && <form onSubmit={confirmDeletion} className="settings-delete-form">
                            {sole.length > 0 && <label className="settings-check">
                                <input type="checkbox" checked={consent} onChange={(event) => setConsent(event.target.checked)} />
                                <span>I understand that my sole-member workspaces and everything inside them will be deleted permanently.</span>
                            </label>}
                            <label htmlFor="settings-delete-password">Confirm your password</label>
                            <input id="settings-delete-password" type="password" autoComplete="current-password"
                                value={password} onChange={(event) => setPassword(event.target.value)} />
                            <label htmlFor="settings-delete-confirmation">Type <strong>DELETE MY ACCOUNT</strong> to confirm</label>
                            <input id="settings-delete-confirmation" type="text" autoComplete="off" spellCheck="false"
                                value={phrase} onChange={(event) => setPhrase(event.target.value)} />
                            {submitError && <p className="settings-deletion-error" role="alert">{submitError}</p>}
                            <div className="settings-delete-actions">
                                <button type="button" className="settings-cancel" disabled={busy}
                                    onClick={() => { setOpen(false); setPassword(''); setPhrase(''); setConsent(false); setSubmitError('') }}>Cancel</button>
                                <button className="settings-confirm-delete" type="submit" disabled={!canDelete}>
                                    {busy ? <><LoaderCircle size={16} className="dashboard-spin" aria-hidden="true" /> Deleting account…</>
                                        : <><Trash2 size={16} aria-hidden="true" /> Delete my account permanently</>}
                                </button>
                            </div>
                        </form>}
                    </>}
            </div>}
        </section>
        </>}
    </section>
}
