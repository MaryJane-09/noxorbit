import { useRef, useState } from 'react'
import {
    ArrowRight,
    Check,
    Layers3,
    LoaderCircle,
    Pencil,
    Plus,
    ShieldCheck,
    X,
} from 'lucide-react'
import WorkspaceMembers from './WorkspaceMembers'
import './WorkspacesPage.css'

const PURPOSES = [
    { value: 'general', label: 'General', detail: 'Flexible for any team' },
    { value: 'education', label: 'Education', detail: 'Study groups and coursework' },
    { value: 'development', label: 'Development', detail: 'Software and developer teams' },
    { value: 'business', label: 'Business', detail: 'Shared goals and initiatives' },
]

function purposeLabel(value) {
    return PURPOSES.find((item) => item.value === value)?.label || 'General'
}

export default function WorkspacesPage({
    workspaces,
    activeWorkspace,
    loading,
    loadError,
    onCreate,
    onRetry,
    onSelect,
    onUpdate,
    onBackToOverview,
}) {
    const editDialogRef = useRef(null)
    const [editingWorkspace, setEditingWorkspace] = useState(null)
    const [membersWorkspaceID, setMembersWorkspaceID] = useState(null)
    const [editName, setEditName] = useState('')
    const [editPurpose, setEditPurpose] = useState('general')
    const [saving, setSaving] = useState(false)
    const [editError, setEditError] = useState('')

    function startEditing(workspace) {
        if (workspace.role !== 'owner') return
        setEditingWorkspace(workspace)
        setEditName(workspace.name)
        setEditPurpose(workspace.purpose)
        setEditError('')
        editDialogRef.current?.showModal()
    }

    function cancelEditing() {
        if (!saving) editDialogRef.current?.close()
    }

    async function saveWorkspace(event) {
        event.preventDefault()
        if (!editingWorkspace || saving) return
        const name = editName.trim()
        if (!name || [...name].length > 100) {
            setEditError('Choose a name between 1 and 100 characters.')
            return
        }
        setSaving(true)
        setEditError('')
        try {
            await onUpdate(editingWorkspace.id, { name, purpose: editPurpose })
            editDialogRef.current?.close()
        } catch (error) {
            setEditError(error.message || 'Could not save workspace changes.')
        } finally {
            setSaving(false)
        }
    }

    return (
        <div className="workspace-manager">
            <header className="workspace-manager-heading">
                <div>
                    <p className="dashboard-eyebrow">YOUR SPACE TO BUILD</p>
                    <h1>Workspaces<span className="workspace-manager-heading-dot">.</span></h1>
                    <p>One home for every team. Keep your work organized, then enter the space you're building in.</p>
                </div>
                <button type="button" className="dashboard-feature-action workspace-manager-create"
                    onClick={onCreate} disabled={loading || Boolean(loadError)}>
                    <Plus size={18} aria-hidden="true" /> New workspace
                </button>
            </header>

            <section className="workspace-manager-section" aria-labelledby="workspace-manager-title">
                <div className="workspace-manager-section-top">
                    <div>
                        <h2 id="workspace-manager-title">Your workspaces</h2>
                        <p>Only workspaces you belong to appear here.</p>
                    </div>
                    {!loading && !loadError && (
                        <span className="workspace-manager-count">
                            {workspaces.length} {workspaces.length === 1 ? 'workspace' : 'workspaces'}
                        </span>
                    )}
                </div>

                {loading && (
                    <div className="workspace-manager-empty" role="status">
                        <LoaderCircle size={26} className="dashboard-spin" aria-hidden="true" />
                        <p>Loading your workspaces…</p>
                    </div>
                )}

                {!loading && loadError && (
                    <div className="workspace-manager-empty" role="alert">
                        <h3>Could not load your workspaces.</h3>
                        <p>{loadError}</p>
                        <button type="button" className="dashboard-feature-action" onClick={onRetry}>
                            Try again <ArrowRight size={17} aria-hidden="true" />
                        </button>
                    </div>
                )}

                {!loading && !loadError && workspaces.length === 0 && (
                    <div className="workspace-manager-empty">
                        <span className="workspace-manager-empty-icon"><Layers3 size={30} aria-hidden="true" /></span>
                        <h3>Start with one shared space.</h3>
                        <p>Give your team a name and purpose. You can add more workspaces later.</p>
                        <button type="button" className="dashboard-feature-action" onClick={onCreate}>
                            <Plus size={17} aria-hidden="true" /> Create your first workspace
                        </button>
                    </div>
                )}

                {!loading && !loadError && workspaces.length > 0 && (
                    <div className="workspace-manager-grid">
                        {workspaces.map((space) => {
                            const isActive = activeWorkspace?.id === space.id
                            return (
                                <article key={space.id}
                                    className={`workspace-manager-card${isActive ? ' workspace-manager-card--active' : ''}`}>
                                    <div className="workspace-manager-card-top">
                                        <span className="workspace-manager-card-icon"><Layers3 size={23} aria-hidden="true" /></span>
                                        {isActive ? <span className="workspace-manager-active"><Check size={13} /> Current</span> : null}
                                    </div>
                                    <h3>{space.name}</h3>
                                    <p className="workspace-manager-meta">
                                        {purposeLabel(space.purpose)} <span aria-hidden="true">·</span> {space.role === 'owner' ? 'Owner' : 'Member'}
                                    </p>
                                    <div className="workspace-manager-card-footer">
                                        <button type="button" className="workspace-manager-open"
                                            onClick={() => { onSelect(space.id); onBackToOverview() }}>
                                            Open overview <ArrowRight size={17} aria-hidden="true" />
                                        </button>
                                        <button type="button" className="workspace-manager-edit"
                                            onClick={() => setMembersWorkspaceID((current) => current === space.id ? null : space.id)}>
                                            {membersWorkspaceID === space.id ? 'Hide members' : 'View members'}
                                        </button>
                                        {space.role === 'owner' && (
                                            <button type="button" className="workspace-manager-edit"
                                                onClick={() => startEditing(space)}
                                                aria-label={`Edit settings for ${space.name}`}>
                                                <Pencil size={17} aria-hidden="true" /> Edit
                                            </button>
                                        )}
                                    </div>
                                </article>
                            )
                        })}
                    </div>
                )}
            </section>

            {membersWorkspaceID && workspaces.some((space) => space.id === membersWorkspaceID) && (
                <WorkspaceMembers key={membersWorkspaceID} workspace={workspaces.find((space) => space.id === membersWorkspaceID)} />
            )}

            <div className="workspace-manager-help">
                <ShieldCheck size={19} aria-hidden="true" />
                <p>Workspace owners can manage invitations and members from View members. Projects and tasks are available in their dedicated sections.</p>
            </div>

            <dialog ref={editDialogRef} className="dashboard-workspace-dialog"
                aria-labelledby="workspace-manager-edit-title"
                onCancel={(event) => { if (saving) event.preventDefault() }}
                onClose={() => { setEditingWorkspace(null); setEditError('') }}>
                <div className="dashboard-dialog-header">
                    <span className="dashboard-dialog-icon"><Pencil size={22} aria-hidden="true" /></span>
                    <button type="button" className="dashboard-dialog-close" onClick={cancelEditing}
                        aria-label="Close edit workspace" disabled={saving}><X size={20} /></button>
                </div>
                <p className="dashboard-eyebrow">WORKSPACE SETTINGS</p>
                <h2 id="workspace-manager-edit-title">Edit workspace.</h2>
                <p className="dashboard-dialog-description">Update this workspace's name or purpose. These changes will be saved for everyone who belongs to it.</p>
                <form className="dashboard-workspace-form" onSubmit={saveWorkspace}>
                    <label htmlFor="workspace-manager-edit-name">Workspace name</label>
                    <input id="workspace-manager-edit-name" type="text" value={editName}
                        onChange={(event) => setEditName(event.target.value)} maxLength={100}
                        required disabled={saving} autoComplete="off" />
                    <label htmlFor="workspace-manager-edit-purpose">Workspace purpose</label>
                    <select id="workspace-manager-edit-purpose" value={editPurpose} disabled={saving}
                        onChange={(event) => setEditPurpose(event.target.value)}>
                        {PURPOSES.map((purpose) => <option key={purpose.value} value={purpose.value}>{purpose.label}</option>)}
                    </select>
                    <p className="dashboard-form-hint">
                        {PURPOSES.find((purpose) => purpose.value === editPurpose)?.detail}
                    </p>
                    {editError && <p className="dashboard-form-error" role="alert">{editError}</p>}
                    <div className="dashboard-dialog-actions">
                        <button type="button" className="dashboard-dialog-cancel" onClick={cancelEditing} disabled={saving}>Cancel</button>
                        <button type="submit" className="dashboard-feature-action dashboard-dialog-submit"
                            disabled={saving || !editName.trim()}>
                            {saving ? <><LoaderCircle size={16} className="dashboard-spin" /> Saving…</>
                                : <><Check size={16} /> Save changes</>}
                        </button>
                    </div>
                </form>
            </dialog>
        </div>
    )
}
