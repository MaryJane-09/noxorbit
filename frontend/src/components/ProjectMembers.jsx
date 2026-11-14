import { useEffect, useState } from 'react'
import { Crown, UserPlus, Users, X } from 'lucide-react'
import { addProjectMember, fetchProjectMembers, fetchWorkspaceMembers, removeProjectMember, updateProjectMemberRole } from '../api'
import './ProjectMembers.css'

export default function ProjectMembers({ workspace, project, user, onMembersChanged }) {
    const [members, setMembers] = useState([])
    const [workspaceMembers, setWorkspaceMembers] = useState([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState('')
    const [notice, setNotice] = useState('')
    const [refresh, setRefresh] = useState(0)
    const [pending, setPending] = useState(false)
    const [target, setTarget] = useState('')
    const [role, setRole] = useState('member')
    const workspaceOwner = workspace.role === 'owner'
    const isLead = members.some((member) => member.user_id === user?.id && member.role === 'lead')
    const canManage = workspaceOwner || isLead

    useEffect(() => {
        const controller = new AbortController()
        setLoading(true)
        Promise.all([
            fetchProjectMembers(workspace.id, project.id, { signal: controller.signal }),
            fetchWorkspaceMembers(workspace.id, { signal: controller.signal }),
        ]).then(([projectMembers, allMembers]) => {
            if (controller.signal.aborted) return
            setMembers(projectMembers)
            setWorkspaceMembers(allMembers)
            setError('')
        }).catch((err) => {
            if (!controller.signal.aborted) setError(err.message || 'Could not load project members.')
        }).finally(() => {
            if (!controller.signal.aborted) setLoading(false)
        })
        return () => controller.abort()
    }, [workspace.id, project.id, refresh])

    const available = workspaceMembers.filter((person) => !members.some((m) => m.user_id === person.user_id))
    async function add(event) {
        event.preventDefault()
        if (!target || pending) return
        setPending(true)
        setError('')
        setNotice('')
        try {
            await addProjectMember(workspace.id, project.id, target, workspaceOwner ? role : 'member')
            setTarget('')
            setRole('member')
            setNotice('Participant added.')
            setRefresh((v) => v + 1)
            onMembersChanged?.()
        } catch (err) {
            setError(err.message || 'Could not add participant.')
        } finally {
            setPending(false)
        }
    }
    async function changeRole(person, nextRole) {
        if (pending || !workspaceOwner || person.role === nextRole) return
        setPending(true)
        setError('')
        setNotice('')
        try {
            await updateProjectMemberRole(workspace.id, project.id, person.user_id, nextRole)
            setNotice('Role updated.')
            setRefresh((v) => v + 1)
        } catch (err) {
            setError(err.message || 'Could not update role.')
        } finally {
            setPending(false)
        }
    }
    async function remove(person) {
        if (pending || !canManage || (!workspaceOwner && person.role === 'lead')) return
        if (!window.confirm(`Remove ${person.name} from this project? Their project task assignments will be cleared.`)) return
        setPending(true)
        setError('')
        setNotice('')
        try {
            await removeProjectMember(workspace.id, project.id, person.user_id)
            setNotice('Participant removed.')
            setRefresh((v) => v + 1)
            onMembersChanged?.()
        } catch (err) {
            setError(err.message || 'Could not remove participant.')
        } finally {
            setPending(false)
        }
    }

    return <section className="project-members" aria-labelledby="project-members-title">
        <div className="project-members-heading">
            <div><h2 id="project-members-title"><Users size={20} aria-hidden="true" /> Project members</h2>
                <p>Workspace members can still view this project. Participants can be assigned its tasks.</p></div>
            <span className="project-members-count">{members.length} participant{members.length === 1 ? '' : 's'}</span>
        </div>
        {loading ? <p role="status">Loading participants…</p> : <>
            {error && <p className="dashboard-form-error" role="alert">{error}</p>}
            {notice && <p className="project-members-success" role="status">{notice}</p>}
            {members.length === 0 && <p className="project-members-empty">No participants yet. Add workspace members to get started.</p>}
            <div className="project-members-list">
                {members.map((person) => <div className="project-members-person" key={person.user_id}>
                    <span className="project-members-initials" aria-hidden="true">{person.name.trim().charAt(0).toUpperCase()}</span>
                    <div className="project-members-person-name"><strong>{person.name}{person.user_id === user?.id ? ' (you)' : ''}</strong>
                        <small>{person.role === 'lead' ? 'Project lead' : 'Project member'}</small></div>
                    {workspaceOwner ? <label className="project-members-role">
                        <span className="project-members-sr-only">Role for {person.name}</span>
                        <select value={person.role} onChange={(e) => changeRole(person, e.target.value)} disabled={pending || project.status !== 'active'}>
                            <option value="member">Member</option><option value="lead">Lead</option>
                        </select>
                    </label> : person.role === 'lead' ? <Crown size={17} aria-label="Project lead" /> : null}
                    {canManage && (workspaceOwner || person.role === 'member') && <button type="button"
                        className="project-members-remove" onClick={() => remove(person)}
                        disabled={pending || project.status !== 'active'} aria-label={`Remove ${person.name} from project`}>
                        <X size={17} aria-hidden="true" />
                    </button>}
                </div>)}
            </div>
            {canManage && project.status === 'active' && available.length > 0 && <form className="project-members-add" onSubmit={add}>
                <label htmlFor="project-member-select">Add a workspace member</label>
                <select id="project-member-select" value={target} onChange={(e) => setTarget(e.target.value)} required disabled={pending}>
                    <option value="">Select a teammate</option>
                    {available.map((person) => <option key={person.user_id} value={person.user_id}>{person.name}</option>)}
                </select>
                {workspaceOwner && <label className="project-members-add-role">
                    <span>Role</span>
                    <select value={role} onChange={(e) => setRole(e.target.value)} disabled={pending}>
                        <option value="member">Member</option><option value="lead">Lead</option>
                    </select>
                </label>}
                <button type="submit" disabled={!target || pending}><UserPlus size={16} aria-hidden="true" />
                    {pending ? 'Adding…' : 'Add member'}</button>
            </form>}
        </>}
    </section>
}
