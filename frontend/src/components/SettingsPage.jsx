import { useEffect, useState } from 'react'
import { AlertTriangle, LoaderCircle, RefreshCw, ShieldAlert, Trash2, UserRound, KeyRound, Camera, Mail, MonitorSmartphone, LogOut } from 'lucide-react'
import { changeAccountPassword, deleteMyAccount, fetchAccountDeletionImpact, saveProfileName, saveMyAvatar, removeMyAvatar, requestEmailChange, confirmEmailChange, fetchActiveSessions, revokeSession, revokeOtherSessions } from '../api'
import './SettingsPage.css'
import AvatarCropper from './AvatarCropper'

export default function SettingsPage({ user, avatarURL, onAvatarUpdated, onEmailChanged, onAccountDeleted, onProfileUpdated, onPasswordChanged }) {
    const [tab, setTab] = useState('profile')
    const [sessions, setSessions] = useState([])
    const [sessionsLoading, setSessionsLoading] = useState(false)
    const [sessionsError, setSessionsError] = useState('')
    const [sessionsMessage, setSessionsMessage] = useState('')
    const [sessionsBusy, setSessionsBusy] = useState(false)
    const [sessionsRevision, setSessionsRevision] = useState(0)
    const [avatarBusy, setAvatarBusy] = useState(false)
    const [avatarFile, setAvatarFile] = useState(null)
    const [avatarError, setAvatarError] = useState('')
    const [avatarNotice, setAvatarNotice] = useState('')
    const [emailFormOpen, setEmailFormOpen] = useState(false)
    const [newEmail, setNewEmail] = useState('')
    const [emailPassword, setEmailPassword] = useState('')
    const [emailCode, setEmailCode] = useState('')
    const [emailRequested, setEmailRequested] = useState(false)
    const [emailBusy, setEmailBusy] = useState(false)
    const [emailError, setEmailError] = useState('')
    const [emailNotice, setEmailNotice] = useState('')
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

    useEffect(() => {
        if (tab !== 'security') return undefined
        let live = true
        fetchActiveSessions()
            .then((data) => {
                if (!live) return
                setSessions(Array.isArray(data.sessions) ? data.sessions : [])
                setSessionsError('')
            })
            .catch((error) => { if (live) setSessionsError(error.message || 'Could not load sessions.') })
            .finally(() => { if (live) setSessionsLoading(false) })
        return () => { live = false }
    }, [tab, sessionsRevision])

    async function signOutSession(item) {
        if (sessionsBusy) return
        setSessionsBusy(true)
        setSessionsError('')
        setSessionsMessage('')
        try {
            await revokeSession(item.id)
            if (item.current) { onPasswordChanged(); return }
            setSessionsMessage('Session signed out.')
            setSessionsLoading(true)
            setSessionsRevision((v) => v + 1)
        } catch (error) {
            setSessionsError(error.message || 'Could not revoke session.')
        } finally {
            setSessionsBusy(false)
        }
    }

    async function signOutOthers() {
        if (sessionsBusy) return
        setSessionsBusy(true)
        setSessionsError('')
        setSessionsMessage('')
        try {
            const result = await revokeOtherSessions()
            setSessionsMessage(`${result.revoked || 0} other session(s) signed out.`)
            setSessionsLoading(true)
            setSessionsRevision((v) => v + 1)
        } catch (error) {
            setSessionsError(error.message || 'Could not revoke other sessions.')
        } finally {
            setSessionsBusy(false)
        }
    }

    const sole = state.impact?.sole_owned || []
    const shared = state.impact?.shared_owned || []
    const blocked = shared.length > 0
    const canDelete = !busy && !state.loading && !state.error && !blocked
        && password.length > 0 && phrase === 'DELETE MY ACCOUNT'
        && (sole.length === 0 || consent)

    function changeAvatar(event) {
        const file = event.target.files?.[0]
        event.target.value = ''
        if (!file || avatarBusy) return
        setAvatarError('')
        setAvatarNotice('')
        if (!['image/png', 'image/jpeg'].includes(file.type) || file.size > 10 * 1024 * 1024) {
            setAvatarError('Choose a PNG or JPEG picture under 10 MB.')
            return
        }
        setAvatarFile(file)
    }

    async function saveCroppedAvatar(file) {
        if (avatarBusy) return
        setAvatarBusy(true)
        setAvatarError('')
        try {
            await saveMyAvatar(file)
            setAvatarFile(null)
            onAvatarUpdated()
            setAvatarNotice('Picture updated.')
        } catch (error) {
            setAvatarError(error.message || 'Could not save picture.')
            throw error
        } finally {
            setAvatarBusy(false)
        }
    }

    async function clearAvatar() {
        if (avatarBusy) return
        setAvatarBusy(true)
        setAvatarError('')
        setAvatarNotice('')
        try {
            await removeMyAvatar()
            onAvatarUpdated()
            setAvatarNotice('Profile picture removed.')
        } catch (error) {
            setAvatarError(error.message || 'Could not remove the picture.')
        } finally {
            setAvatarBusy(false)
        }
    }

    async function startEmailChange(event) {
        event.preventDefault()
        if (emailBusy) return
        setEmailBusy(true)
        setEmailError('')
        setEmailNotice('')
        try {
            await requestEmailChange(newEmail.trim(), emailPassword)
            setEmailPassword('')
            setEmailCode('')
            setEmailRequested(true)
            setEmailNotice('Check your new inbox for the six-character verification code. It expires in five minutes.')
        } catch (error) {
            setEmailError(error.message || 'Could not send the verification code.')
        } finally {
            setEmailBusy(false)
        }
    }

    async function finishEmailChange(event) {
        event.preventDefault()
        if (emailBusy) return
        setEmailBusy(true)
        setEmailError('')
        try {
            await confirmEmailChange(emailCode.trim())
            onEmailChanged()
        } catch (error) {
            setEmailError(error.message || 'Could not verify the code.')
        } finally {
            setEmailBusy(false)
        }
    }

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

    return <>
        {avatarFile && <AvatarCropper file={avatarFile}
            onCancel={() => { if (!avatarBusy) setAvatarFile(null) }}
            onSave={saveCroppedAvatar} saving={avatarBusy} />}
        <section className="settings-page" aria-labelledby="settings-heading">
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
                aria-current={tab === 'security' ? 'page' : undefined} onClick={() => { setSessionsLoading(true); setTab('security'); setSessionsRevision((v) => v + 1) }}>
                <KeyRound size={18} aria-hidden="true" /> Security
            </button>
        </nav>
        {tab === 'profile' && <section className="settings-account" aria-labelledby="settings-account-heading">
            <div className="settings-section-title">
                <UserRound size={21} aria-hidden="true" />
                <h2 id="settings-account-heading">My profile</h2>
            </div>
            <div className="settings-avatar-editor">
                <span className="settings-avatar-preview">
                    {avatarURL ? <img src={avatarURL} alt="Your profile picture" /> : <UserRound size={35} aria-hidden="true" />}
                </span>
                <div>
                    <strong>Profile picture</strong>
                    <p className="settings-help">PNG or JPEG · Max 10 MB</p>
                    <div className="settings-avatar-actions">
                        <label className="settings-avatar-upload" htmlFor="settings-avatar-input">
                            <Camera size={16} aria-hidden="true" /> {avatarBusy ? 'Saving…' : 'Upload picture'}
                        </label>
                        <input id="settings-avatar-input" className="settings-avatar-input"
                            type="file" accept="image/png,image/jpeg"
                            onChange={changeAvatar} disabled={avatarBusy} />
                        {avatarURL && <button type="button" onClick={clearAvatar} disabled={avatarBusy}>Remove picture</button>}
                    </div>
                </div>
            </div>
            {avatarError && <p role="alert" className="settings-deletion-error">{avatarError}</p>}
            {avatarNotice && <p role="status" className="settings-success">{avatarNotice}</p>}
            <p className="settings-help">Your name is visible to teammates across Noxorbit.</p>
            <form onSubmit={saveName} className="settings-edit-form">
                <label htmlFor="settings-display-name">Full name</label>
                <input id="settings-display-name" value={displayName} autoComplete="name" maxLength={100}
                    onChange={(event) => setDisplayName(event.target.value)} />
                <label htmlFor="settings-email-display">Current email address</label>
                <input id="settings-email-display" value={user.email} readOnly />
                <p className="settings-help">Changing your email requires your password and a code sent to the new address.</p>
                {profileError && <p role="alert" className="settings-deletion-error">{profileError}</p>}
                {profileMessage && <p role="status" className="settings-success">{profileMessage}</p>}
                <div className="settings-edit-actions">
                    <button type="submit" disabled={savingName || displayName.trim() === user.name}>
                        {savingName ? 'Saving…' : 'Save changes'}
                    </button>
                </div>
            </form>
            <div className="settings-email-change">
                <div className="settings-email-heading">
                    <Mail size={19} aria-hidden="true" />
                    <strong>Change email address</strong>
                    <button type="button" onClick={() => {
                        setEmailFormOpen((value) => !value)
                        setEmailError('')
                        setEmailNotice('')
                        setEmailRequested(false)
                        setNewEmail('')
                        setEmailPassword('')
                        setEmailCode('')
                    }} disabled={emailBusy}>
                        {emailFormOpen ? 'Cancel' : 'Change email'}
                    </button>
                </div>
                {emailFormOpen && (!emailRequested ? <form className="settings-edit-form" onSubmit={startEmailChange}>
                    <label htmlFor="settings-new-email">New email address</label>
                    <input id="settings-new-email" type="email" autoComplete="email"
                        value={newEmail} onChange={(event) => setNewEmail(event.target.value)} required />
                    <label htmlFor="settings-email-password">Current password</label>
                    <input id="settings-email-password" type="password" autoComplete="current-password"
                        value={emailPassword} onChange={(event) => setEmailPassword(event.target.value)} required />
                    <div className="settings-edit-actions">
                        <button type="submit" disabled={emailBusy || !newEmail || !emailPassword}>
                            {emailBusy ? 'Sending…' : 'Send verification code'}
                        </button>
                    </div>
                </form> : <form className="settings-edit-form" onSubmit={finishEmailChange}>
                    <p className="settings-help">Enter the code sent to <strong>{newEmail}</strong>. Your current email remains active until verification succeeds.</p>
                    <label htmlFor="settings-email-code">Verification code</label>
                    <input id="settings-email-code" type="text" inputMode="text" autoComplete="one-time-code"
                        maxLength={6} value={emailCode} onChange={(event) => setEmailCode(event.target.value)} required />
                    <div className="settings-edit-actions">
                        <button type="submit" disabled={emailBusy || emailCode.trim().length !== 6}>
                            {emailBusy ? 'Verifying…' : 'Verify and change email'}
                        </button>
                    </div>
                </form>)}
                {emailError && <p role="alert" className="settings-deletion-error">{emailError}</p>}
                {emailNotice && <p role="status" className="settings-success">{emailNotice}</p>}
            </div>
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
            <section className="settings-account" aria-labelledby="settings-sessions-heading">
                <div className="settings-section-title"><MonitorSmartphone size={21} aria-hidden="true" />
                    <h2 id="settings-sessions-heading">Active sessions</h2>
                </div>
                <p className="settings-help">Review where you're signed in and sign out sessions you don't recognize. Older sessions may not have device details.</p>
                {sessionsLoading && <p className="settings-help" role="status">Loading sessions…</p>}
                {sessionsError && <p role="alert" className="settings-deletion-error">{sessionsError}</p>}
                {sessionsMessage && <p role="status" className="settings-success">{sessionsMessage}</p>}
                {!sessionsLoading && sessions.map((item) => <div className="settings-session-row" key={item.id}>
                    <MonitorSmartphone size={21} aria-hidden="true" />
                    <div className="settings-session-details">
                        <strong>{item.user_agent || 'Unknown device'} {item.current && <span className="settings-session-current">Current</span>}</strong>
                        <span>Signed in {new Date(item.created_at).toLocaleString()} · Expires {new Date(item.expires_at).toLocaleDateString()}</span>
                    </div>
                    <button type="button" disabled={sessionsBusy} onClick={() => signOutSession(item)}>
                        <LogOut size={15} aria-hidden="true" /> Sign out
                    </button>
                </div>)}
                {!sessionsLoading && !sessionsError && sessions.length === 0 && <p className="settings-help">No active sessions found.</p>}
                {sessions.filter((item) => !item.current).length > 0 && <div className="settings-session-actions">
                    <button type="button" disabled={sessionsBusy} onClick={signOutOthers}>
                        Sign out all other sessions
                    </button>
                </div>}
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
    </>
}
