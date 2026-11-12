import { useEffect, useState } from 'react'
import { LoaderCircle, Users } from 'lucide-react'
import { createWorkspaceInvite, fetchWorkspaceMembers } from '../api'

export default function WorkspaceMembers({ workspace }) {
    const [state, setState] = useState({ id: null, members: [], error: '' })
    const [reload, setReload] = useState(0)
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
    async function generateLink() {
        if (invite.creating) return
        setInvite({ creating: true, link: '', error: '', copied: false })
        try {
            const result = await createWorkspaceInvite(workspace.id)
            const url = `${window.location.origin}${window.location.pathname}#invite/${encodeURIComponent(result.token)}`
            setInvite({ creating: false, link: url, error: '', copied: false })
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
                        </div>
                    ))}
                </div>}
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
        </section>
    )
}
