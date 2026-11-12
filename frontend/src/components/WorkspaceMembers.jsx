import { useEffect, useState } from 'react'
import { LoaderCircle, Users } from 'lucide-react'
import { fetchWorkspaceMembers } from '../api'

export default function WorkspaceMembers({ workspace }) {
    const [state, setState] = useState({ id: null, members: [], error: '' })
    const [reload, setReload] = useState(0)
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
            {workspace.role === 'owner' && <p className="dashboard-form-hint">
                Secure invitation links and email invitations are the next stage. This list currently shows existing members.
            </p>}
        </section>
    )
}
