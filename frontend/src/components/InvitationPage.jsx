import { useState } from 'react'
import { Check, LoaderCircle, Users } from 'lucide-react'
import { acceptWorkspaceInvite } from '../api'

export default function InvitationPage({ token, onAccepted }) {
    const [busy, setBusy] = useState(false)
    const [error, setError] = useState('')

    async function accept() {
        if (busy || !token) return
        setBusy(true)
        setError('')
        try {
            const result = await acceptWorkspaceInvite(token)
            onAccepted(result.workspace_id)
        } catch (err) {
            setError(err.message || 'Could not accept invitation.')
            setBusy(false)
        }
    }

    return <section className="projects-message" aria-labelledby="invite-heading">
        <Users size={28} aria-hidden="true" />
        <p className="dashboard-eyebrow">BETTER WORK HAPPENS TOGETHER</p>
        <h2 id="invite-heading">Join a workspace.</h2>
        <p>This invitation will add your account as a member of the inviting team.
            Only accept invitations from people you trust.</p>
        {error && <p role="alert" className="dashboard-form-error">{error}</p>}
        <button type="button" className="dashboard-feature-action" disabled={busy || !token} onClick={accept}>
            {busy ? <><LoaderCircle size={16} className="dashboard-spin" /> Joining…</>
                : <><Check size={16} /> Accept invitation</>}
        </button>
    </section>
}
