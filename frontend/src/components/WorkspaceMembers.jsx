import { useEffect, useRef, useState } from 'react'
import { LoaderCircle, Users, Trash2, X } from 'lucide-react'
import { createWorkspaceInvite, fetchWorkspaceMembers, fetchWorkspaceInvites, removeWorkspaceMember, revokeWorkspaceInvite } from '../api'

export default function WorkspaceMembers({ workspace }) {
    const [state, setState] = useState({ id: null, members: [], error: '' })
    const [reload, setReload] = useState(0)
    const [invites, setInvites] = useState({ loading: false, items: [], error: '' })
    const [confirm, setConfirm] = useState(null)
    const [working, setWorking] = useState(false)
    const [actionError, setActionError] = useState('')
    const confirmRef = useRef(null)
    const [invite, setInvite] = useState({ creating: false, link: '', error: '', copied: false })
    useEffect(() => {
        const controller = new AbortController()
        fetchWorkspaceMembers(workspace.id, { signal: controller.signal })
            .then((members) => {
                if (!controller.signal.aborted) setState({ id: workspace.id, members, error: '' })
            })
            .catch((err) => {
                if (!controller.signal.aborted) setState({ id: workspace.id, members: [], error: err.message })
            })
        return () => controller.abort()
    }, [workspace.id, reload])
    useEffect(() => {
        if (workspace.role !== 'owner') return
        const controller = new AbortController()
        fetchWorkspaceInvites(workspace.id, { signal: controller.signal })
            .then((items) => {
                if (!controller.signal.aborted) setInvites({ loading: false, items, error: '' })
            })
            .catch((err) => {
                if (!controller.signal.aborted) setInvites({ loading: false, items: [], error: err.message })
            })
        return () => controller.abort()
    }, [workspace.id, workspace.role, reload])

    useEffect(() => {
        const dialog = confirmRef.current
        if (confirm && dialog && !dialog.open) dialog.showModal()
        if (!confirm && dialog?.open) dialog.close()
    }, [confirm])

    async function runConfirmed(event) {
        event.preventDefault()
        if (!confirm || working) return
        setWorking(true)
        setActionError('')
        try {
            if (confirm.type === 'member') {
                await removeWorkspaceMember(workspace.id, confirm.id)
            } else {
                await revokeWorkspaceInvite(workspace.id, confirm.id)
            }
            setConfirm(null)
            setReload((value) => value + 1)
        } catch (err) {
            setActionError(err.message || 'Action failed.')
        } finally {
            setWorking(false)
        }
    }

    function cancelAction() {
        if (!working) { setConfirm(null); setActionError('') }
    }

    async function generateLink() {
        if (invite.creating) return
        setInvite({ creating: true, link: '', error: '', copied: false })
        try {
            const result = await createWorkspaceInvite(workspace.id)
            const url = `${window.location.origin}${window.location.pathname}#invite/${encodeURIComponent(result.token)}`
            setInvite({ creating: false, link: url, error: '', copied: false })
            setReload((value) => value + 1)
        } catch (err) {
            setInvite({ creating: false, link: '', error: err.message, copied: false })
        }
    }
    async function copyLink() {
        try {
            await navigator.clipboard.writeText(invite.link)
            setInvite((current) => ({ ...current, copied: true }))
        } catch {
            setInvite((current) => ({ ...current, error: 'Copy failed. Select and copy the link manually.' }))
        }
    }

    return (
        <section className="workspace-manager-section" aria-label={`Members of ${workspace.name}`}>
            <div className="workspace-manager-section-top">
                <div>
                    <h2><Users size={19} aria-hidden="true" /> Members of {workspace.name}</h2>
                    <p>Only people who belong to this workspace appear here.</p>
                </div>
                {state.id === workspace.id && !state.error &&
                    <span className="workspace-manager-count">{state.members.length} members</span>}
            </div>
            {state.id !== workspace.id && <p role="status"><LoaderCircle className="dashboard-spin" size={17} /> Loading members…</p>}
            {state.id === workspace.id && state.error && <div role="alert">
                <p className="dashboard-form-error">{state.error}</p>
                <button type="button" className="dashboard-dialog-cancel" onClick={() => setReload((n) => n+1)}>Retry</button>
            </div>}
            {state.id === workspace.id && !state.error &&
                <div className="workspace-members-list">
                    {state.members.map((member) => (
                        <div className="workspace-member-row" key={member.user_id}>
                            <strong>{member.name}</strong>
                            <span>{member.role === 'owner' ? 'Owner' : 'Member'}</span>
                            {workspace.role === 'owner' && member.role !== 'owner' && (
                                <button type="button" className="workspace-member-remove"
                                    onClick={() => setConfirm({ type: 'member', id: member.user_id, label: member.name })}>
                                    <Trash2 size={14} aria-hidden="true" /> Remove
                                </button>
                            )}
                        </div>
                    ))}
                </div>}
            {workspace.role === 'owner' && <section className="workspace-invite-history" aria-label="Invitation history">
                <h3>Invitation history</h3>
                {invites.error && <p className="dashboard-form-error" role="alert">{invites.error}</p>}
                {invites.items.length === 0 && !invites.error && <p className="dashboard-form-hint">No invitations yet.</p>}
                {invites.items.map((item) => {
                    const expired = new Date(item.expires_at).getTime() <= Date.now()
                    const used = item.use_count >= item.max_uses
                    const revoked = Boolean(item.revoked_at)
                    const status = revoked ? 'Revoked' : used ? 'Used' : expired ? 'Expired' : 'Active'
                    return <div key={item.id} className="workspace-invite-entry">
                        <div><strong>{item.kind === 'email' ? 'Email invitation' : 'Invitation link'}</strong>
                            <span>{status} · {item.use_count}/{item.max_uses} uses · Expires {new Date(item.expires_at).toLocaleDateString()}</span></div>
                        {!revoked && !used && !expired && <button type="button" className="workspace-member-remove"
                            onClick={() => setConfirm({ type: 'invite', id: item.id, label: 'this invitation link' })}>
                            Revoke
                        </button>}
                    </div>
                })}
            </section>}
            {workspace.role === 'owner' && <div className="workspace-invite-tools">
                <h3>Invite a teammate</h3>
                <p className="dashboard-form-hint">Create a link valid for seven days and one new member. Send it privately to the person you want to invite.</p>
                <button type="button" className="dashboard-feature-action" disabled={invite.creating}
                    onClick={generateLink}>{invite.creating ? 'Generating…' : 'Generate invitation link'}</button>
                {invite.link && <div className="workspace-invite-link">
                    <label htmlFor="workspace-invite-url">Invitation link</label>
                    <input id="workspace-invite-url" readOnly value={invite.link} onFocus={(e) => e.target.select()} />
                    <button type="button" className="dashboard-dialog-cancel" onClick={copyLink}>
                        {invite.copied ? 'Copied' : 'Copy link'}
                    </button>
                </div>}
                {invite.error && <p className="dashboard-form-error" role="alert">{invite.error}</p>}
            </div>}
            <dialog ref={confirmRef} className="dashboard-workspace-dialog workspace-member-confirm"
                aria-labelledby="member-confirm-heading"
                onCancel={(event) => { if (working) event.preventDefault(); else cancelAction() }}
                onClose={() => { if (!working) { setConfirm(null); setActionError('') } }}>
                <div className="dashboard-dialog-header">
                    <span className="dashboard-dialog-icon"><Trash2 size={20} aria-hidden="true" /></span>
                    <button type="button" className="dashboard-dialog-close" disabled={working}
                        aria-label="Cancel action" onClick={cancelAction}><X size={20} /></button>
                </div>
                <p className="dashboard-eyebrow">CONFIRM YOUR ACTION</p>
                <h2 id="member-confirm-heading">{confirm?.type === 'member' ? 'Remove member?' : 'Revoke invitation?'}</h2>
                <p className="dashboard-dialog-description">
                    {confirm?.type === 'member'
                        ? <>Remove <strong>{confirm.label}</strong> from this workspace? They will lose access, and any tasks assigned to them will become unassigned.</>
                        : 'This invitation will no longer allow anyone new to join this workspace.'}
                </p>
                {actionError && <p className="dashboard-form-error" role="alert">{actionError}</p>}
                <form onSubmit={runConfirmed} className="dashboard-dialog-actions">
                    <button type="button" className="dashboard-dialog-cancel" disabled={working} onClick={cancelAction}>Cancel</button>
                    <button type="submit" className="workspace-member-danger" disabled={working}>
                        {working ? 'Saving…' : confirm?.type === 'member' ? 'Remove member' : 'Revoke invitation'}
                    </button>
                </form>
            </dialog>
        </section>
    )
}
