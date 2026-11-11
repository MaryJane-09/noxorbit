import { useEffect, useRef, useState } from 'react'
import {
    Archive,
    ArrowRight,
    ArrowUpRight,
    CalendarDays,
    Check,
    FolderKanban,
    Layers3,
    LoaderCircle,
    Pencil,
    Plus,
    RotateCcw,
    X,
} from 'lucide-react'
import { changeProjectStatus, createProject, fetchProjects, updateProject } from '../api'
import TasksBoard from './TasksBoard'
import './ProjectsPage.css'

export function ProjectOverviewCard({ workspaceID, onOpen }) {
    const [state, setState] = useState({ workspaceID: null, projects: [], error: '' })
    useEffect(() => {
        if (!workspaceID) return
        const controller = new AbortController()
        fetchProjects(workspaceID, { signal: controller.signal })
            .then((items) => {
                if (!controller.signal.aborted) {
                    setState({ workspaceID, projects: items, error: '' })
                }
            })
            .catch((error) => {
                if (!controller.signal.aborted) {
                    setState({ workspaceID, projects: [], error: error.message })
                }
            })
        return () => controller.abort()
    }, [workspaceID])

    const ready = workspaceID && state.workspaceID === workspaceID
    const active = ready ? state.projects.filter((project) => project.status === 'active').length : 0
    const archived = ready ? state.projects.length - active : 0

    return (
        <section className="dashboard-empty-card" aria-labelledby="dashboard-projects-heading">
            <div className="dashboard-empty-top">
                <span className="dashboard-tile-icon"><FolderKanban size={23} aria-hidden="true" /></span>
                <ArrowUpRight size={19} aria-hidden="true" />
            </div>
            <h3 id="dashboard-projects-heading">Projects</h3>
            <p>{!workspaceID ? 'Create a workspace to give your projects a home.'
                : !ready ? 'Loading projects…'
                : state.error ? 'Project data is unavailable. Open Projects to retry.'
                : state.projects.length === 0 ? 'No projects yet. Start with one shared goal.'
                : `${active} active ${active === 1 ? 'project' : 'projects'}${archived ? ` · ${archived} archived` : ''} in this workspace.`}</p>
            {workspaceID && (
                <button type="button" className="dashboard-projects-open" onClick={onOpen}>
                    Open Projects <ArrowRight size={15} aria-hidden="true" />
                </button>
            )}
            <span className="dashboard-empty-status">
                {!workspaceID ? 'Not set up yet' : !ready ? 'Loading' : state.error ? 'Unavailable'
                    : state.projects.length === 0 ? 'Ready for your first project' : 'Live workspace data'}
            </span>
        </section>
    )
}

export default function ProjectsPage({ activeWorkspace, workspacesLoading, onOpenWorkspaces, user }) {
    const [selectedProject, setSelectedProject] = useState(null)
    const [data, setData] = useState({ workspaceID: null, items: [], error: '' })
    const [reload, setReload] = useState(0)
    const [tab, setTab] = useState('active')
    const [editor, setEditor] = useState(null)
    const [name, setName] = useState('')
    const [description, setDescription] = useState('')
    const [saving, setSaving] = useState(false)
    const [busyID, setBusyID] = useState(null)
    const [formError, setFormError] = useState('')
    const [actionError, setActionError] = useState('')
    const dialogRef = useRef(null)
    const nameRef = useRef(null)
    const workspaceID = activeWorkspace?.id
    const ready = Boolean(workspaceID && data.workspaceID === workspaceID)

    useEffect(() => {
        if (!workspaceID) return
        const controller = new AbortController()
        fetchProjects(workspaceID, { signal: controller.signal })
            .then((items) => {
                if (!controller.signal.aborted) {
                    setData({ workspaceID, items, error: '' })
                }
            })
            .catch((error) => {
                if (!controller.signal.aborted) {
                    setData({ workspaceID, items: [], error: error.message })
                }
            })
        return () => controller.abort()
    }, [workspaceID, reload])

    function retry() {
        setData({ workspaceID: null, items: [], error: '' })
        setReload((value) => value + 1)
    }

    function openEditor(project = null) {
        setEditor(project || { id: null })
        setName(project?.name || '')
        setDescription(project?.description || '')
        setFormError('')
        dialogRef.current?.showModal()
        nameRef.current?.focus()
    }

    function cancelEditor() {
        if (!saving) dialogRef.current?.close()
    }

    async function save(event) {
        event.preventDefault()
        if (saving || !workspaceID) return
        const cleaned = name.trim()
        if (!cleaned || [...cleaned].length > 100 || [...description].length > 2000) {
            setFormError('Name must be 1–100 characters and description at most 2000.')
            return
        }
        setSaving(true)
        setFormError('')
        try {
            const saved = editor.id
                ? await updateProject(workspaceID, editor.id, { name: cleaned, description })
                : await createProject(workspaceID, { name: cleaned, description })
            setData((current) => current.workspaceID === workspaceID
                ? {
                    ...current,
                    items: editor.id
                        ? current.items.map((item) => item.id === saved.id ? saved : item)
                        : [saved, ...current.items],
                }
                : current)
            setTab(saved.status)
            dialogRef.current?.close()
        } catch (error) {
            setFormError(error.message || 'Could not save your project.')
        } finally {
            setSaving(false)
        }
    }

    async function toggleArchive(project) {
        if (!workspaceID || busyID) return
        setBusyID(project.id)
        setActionError('')
        try {
            const updated = await changeProjectStatus(
                workspaceID, project.id,
                project.status === 'archived' ? 'active' : 'archived'
            )
            setData((current) => current.workspaceID === workspaceID
                ? { ...current, items: current.items.map((item) => item.id === updated.id ? updated : item) }
                : current)
        } catch (error) {
            setActionError(error.message || 'Could not change project status.')
        } finally {
            setBusyID(null)
        }
    }

    const items = ready ? data.items : []
    const visible = items.filter((project) => project.status === tab)
    const count = (status) => items.filter((project) => project.status === status).length
    const isOwner = activeWorkspace?.role === 'owner'
    if (selectedProject && activeWorkspace && selectedProject.workspaceID === activeWorkspace.id) {
        const currentProject = items.find((item) => item.id === selectedProject.id)
        if (currentProject) {
            return <TasksBoard key={`${activeWorkspace.id}:${currentProject.id}`} workspace={activeWorkspace}
                project={currentProject} user={user} onBack={() => setSelectedProject(null)} />
        }
    }

    return (
        <div className="projects-page">
            <header className="projects-page-heading">
                <div>
                    <p className="dashboard-eyebrow">A PLACE FOR EVERY IDEA</p>
                    <h1>Projects<span>.</span></h1>
                    <p>{activeWorkspace
                        ? `Organize meaningful work inside ${activeWorkspace.name}.`
                        : 'Select a workspace to start organizing projects.'}</p>
                </div>
                {activeWorkspace && ready && !data.error && (
                    <button type="button" className="dashboard-feature-action projects-create"
                        onClick={() => openEditor()}>
                        <Plus size={17} aria-hidden="true" /> New project
                    </button>
                )}
            </header>

            {workspacesLoading && !activeWorkspace && (
                <div className="projects-message" role="status">
                    <LoaderCircle size={23} className="dashboard-spin" aria-hidden="true" />
                    <p>Loading your workspaces…</p>
                </div>
            )}
            {!workspacesLoading && !activeWorkspace && (
                <div className="projects-message">
                    <Layers3 size={28} aria-hidden="true" />
                    <h2>Create or choose a workspace first.</h2>
                    <p>Projects belong to workspaces so teams can build together.</p>
                    <button type="button" className="dashboard-feature-action" onClick={onOpenWorkspaces}>
                        View workspaces <ArrowRight size={17} aria-hidden="true" />
                    </button>
                </div>
            )}
            {activeWorkspace && !ready && (
                <div className="projects-message" role="status">
                    <LoaderCircle size={23} className="dashboard-spin" aria-hidden="true" />
                    <p>Loading projects…</p>
                </div>
            )}
            {ready && data.error && (
                <div className="projects-message" role="alert">
                    <h2>Could not load projects.</h2>
                    <p>{data.error}</p>
                    <button type="button" className="dashboard-feature-action" onClick={retry}>
                        Try again <ArrowRight size={17} aria-hidden="true" />
                    </button>
                </div>
            )}

            {ready && !data.error && (
                <>
                    <div className="projects-page-tools">
                        <div className="projects-page-tabs" role="group" aria-label="Project status">
                            <button type="button" className={tab === 'active' ? 'is-active' : ''}
                                aria-pressed={tab === 'active'} onClick={() => setTab('active')}>
                                Active <span>{count('active')}</span>
                            </button>
                            <button type="button" className={tab === 'archived' ? 'is-active' : ''}
                                aria-pressed={tab === 'archived'} onClick={() => setTab('archived')}>
                                Archived <span>{count('archived')}</span>
                            </button>
                        </div>
                        <span className="projects-page-workspace">{activeWorkspace.name}</span>
                    </div>
                    {actionError && <p className="dashboard-form-error projects-action-error" role="alert">{actionError}</p>}
                    {visible.length === 0 ? (
                        <div className="projects-message">
                            <span className="projects-empty-icon">
                                {tab === 'active' ? <FolderKanban size={30} /> : <Archive size={28} />}
                            </span>
                            <h2>{tab === 'active' ? 'Every great project starts somewhere.' : 'Nothing archived yet.'}</h2>
                            <p>{tab === 'active'
                                ? 'Create a project and give your team a shared goal.'
                                : 'Projects archived by an owner will appear here.'}</p>
                            {tab === 'active' && (
                                <button type="button" className="dashboard-feature-action" onClick={() => openEditor()}>
                                    <Plus size={17} aria-hidden="true" /> Create project
                                </button>
                            )}
                        </div>
                    ) : (
                        <div className="projects-grid">
                            {visible.map((project) => (
                                <article key={project.id} className="projects-card">
                                    <div className="projects-card-top">
                                        <span className="projects-card-icon"><FolderKanban size={23} /></span>
                                        <span className="projects-card-status">
                                            {project.status === 'active' ? <Check size={14} /> : <Archive size={14} />}
                                            {project.status}
                                        </span>
                                    </div>
                                    <h2>{project.name}</h2>
                                    <p>{project.description || 'No description yet.'}</p>
                                    <div className="projects-card-bottom">
                                        <span className="projects-card-date">
                                            <CalendarDays size={14} aria-hidden="true" />
                                            {new Date(project.created_at).toLocaleDateString(undefined, {
                                                month: 'short', day: 'numeric', year: 'numeric',
                                            })}
                                        </span>
                                        <div className="projects-card-actions">
                                            <button type="button" onClick={() => setSelectedProject({ id: project.id, workspaceID })}>
                                                <ArrowRight size={16} /> Open board
                                            </button>
                                            <button type="button" onClick={() => openEditor(project)}
                                                aria-label={`Edit ${project.name}`}><Pencil size={16} /> Edit</button>
                                            {isOwner && (
                                                <button type="button" onClick={() => toggleArchive(project)}
                                                    disabled={busyID !== null}
                                                    aria-label={`${project.status === 'active' ? 'Archive' : 'Restore'} ${project.name}`}>
                                                    {busyID === project.id
                                                        ? <LoaderCircle size={16} className="dashboard-spin" />
                                                        : project.status === 'active' ? <Archive size={16} /> : <RotateCcw size={16} />}
                                                    {project.status === 'active' ? 'Archive' : 'Restore'}
                                                </button>
                                            )}
                                        </div>
                                    </div>
                                </article>
                            ))}
                        </div>
                    )}
                </>
            )}

            <dialog className="dashboard-workspace-dialog projects-dialog" ref={dialogRef}
                aria-labelledby="projects-dialog-title"
                onCancel={(event) => { if (saving) event.preventDefault() }}
                onClose={() => { setEditor(null); setFormError('') }}>
                <div className="dashboard-dialog-header">
                    <span className="dashboard-dialog-icon"><FolderKanban size={23} aria-hidden="true" /></span>
                    <button type="button" className="dashboard-dialog-close" onClick={cancelEditor}
                        disabled={saving} aria-label="Close project form"><X size={20} /></button>
                </div>
                <p className="dashboard-eyebrow">SHARED WORK, CLEAR DIRECTION</p>
                <h2 id="projects-dialog-title">{editor?.id ? 'Edit project.' : 'Create a project.'}</h2>
                <p className="dashboard-dialog-description">Projects live inside your selected workspace.</p>
                <form className="dashboard-workspace-form" onSubmit={save}>
                    <label htmlFor="projects-name">Project name</label>
                    <input id="projects-name" ref={nameRef} value={name} autoComplete="off"
                        onChange={(event) => setName(event.target.value)} required maxLength={100} disabled={saving}
                        placeholder="e.g. Semester research project" />
                    <label htmlFor="projects-description">Description (optional)</label>
                    <textarea id="projects-description" rows={4} maxLength={2000} disabled={saving}
                        value={description} onChange={(event) => setDescription(event.target.value)}
                        placeholder="What will your team work on?" />
                    {formError && <p className="dashboard-form-error" role="alert">{formError}</p>}
                    <div className="dashboard-dialog-actions">
                        <button type="button" className="dashboard-dialog-cancel" onClick={cancelEditor} disabled={saving}>Cancel</button>
                        <button type="submit" className="dashboard-feature-action dashboard-dialog-submit"
                            disabled={saving || !name.trim()}>
                            {saving ? <><LoaderCircle size={16} className="dashboard-spin" /> Saving…</>
                                : <><Check size={16} /> {editor?.id ? 'Save changes' : 'Create project'}</>}
                        </button>
                    </div>
                </form>
            </dialog>
        </div>
    )
}
