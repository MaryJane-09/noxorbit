import { useEffect, useRef, useState } from 'react'
import {
    ArrowRight,
    CheckCheck,
    CalendarDays,
    ChevronRight,
    Compass,
    FolderKanban,
    LayoutDashboard,
    LoaderCircle,
    Layers3,
    Lightbulb,
    ListTodo,
    LogOut,
    Trash2,
    Menu,
    Plus,
    Sparkles,
    Settings2,
    Users,
    X,
} from 'lucide-react'
import { API, createWorkspace, fetchWorkspaces, updateWorkspace } from '../api'
import WorkspacesPage from './WorkspacesPage'
import InvitationPage from './InvitationPage'
import MyTasksPage from './MyTasksPage'
import RecycleBinPage from './RecycleBinPage'
import ActivityPage from './ActivityPage'
import SettingsPage from './SettingsPage'
import OverviewTaskCards from './OverviewTaskCards'
import ProjectsPage, { ProjectOverviewCard } from './ProjectsPage'
import './Home.css'

const SETUP_STEPS = [
    {
        number: '01',
        title: 'Create a workspace',
        description: 'Give your team a home and choose how you plan to work together.',
    },
    {
        number: '02',
        title: 'Bring your team together',
        description: 'Invite the people you want to build, study, or create with.',
    },
    {
        number: '03',
        title: 'Turn plans into progress',
        description: 'Add projects, assign tasks, and keep deadlines in one place.',
    },
]

function initialLetters(name) {
    const parts = (name || '').trim().split(/\s+/).filter(Boolean)
    return (parts.length > 1 ? parts[0][0] + parts[parts.length - 1][0] : (parts[0] || 'N').slice(0, 2)).toUpperCase()
}

const WORKSPACE_PURPOSES = [
    { value: 'general', label: 'General', detail: 'A flexible space for any team.' },
    { value: 'education', label: 'Education', detail: 'Study groups, coursework, and clubs.' },
    { value: 'development', label: 'Development', detail: 'Software projects and developer teams.' },
    { value: 'business', label: 'Business', detail: 'Teams, plans, and shared goals.' },
]

function purposeLabel(purpose) {
    return WORKSPACE_PURPOSES.find((choice) => choice.value === purpose)?.label || 'General'
}

function pageFromHash() {
    if (window.location.hash.startsWith('#invite/')) return 'invite'
    if (window.location.hash === '#workspaces') return 'workspaces'
    if (window.location.hash === '#projects') return 'projects'
    if (window.location.hash === '#my-tasks') return 'my-tasks'
    if (window.location.hash === '#recycle-bin') return 'recycle-bin'
    if (window.location.hash === '#activity') return 'activity'
    if (window.location.hash === '#settings') return 'settings'
    return 'overview'
}

export default function Home({ user, onLogout }) {
    const [loggingOut, setLoggingOut] = useState(false)
    const [error, setError] = useState('')
    const [guideOpen, setGuideOpen] = useState(false)
    const [menuOpen, setMenuOpen] = useState(false)
    const [page, setPage] = useState(pageFromHash)
    const [workspaces, setWorkspaces] = useState([])
    const [workspacesLoading, setWorkspacesLoading] = useState(true)
    const [workspaceLoadError, setWorkspaceLoadError] = useState('')
    const [reloadVersion, setReloadVersion] = useState(0)
    const [selectedWorkspaceId, setSelectedWorkspaceId] = useState(null)
    const [createDialogOpen, setCreateDialogOpen] = useState(false)
    const [creatingWorkspace, setCreatingWorkspace] = useState(false)
    const [createError, setCreateError] = useState('')
    const [workspaceName, setWorkspaceName] = useState('')
    const [workspacePurpose, setWorkspacePurpose] = useState('general')
    const dialogRef = useRef(null)
    const nameInputRef = useRef(null)
    const displayName = (user.name || '').trim().split(/\s+/)[0] || 'there'
    const activeWorkspace = workspaces.find((space) => space.id === selectedWorkspaceId) || workspaces[0] || null
    const hasWorkspace = Boolean(activeWorkspace)

    useEffect(() => {
        const handleHashChange = () => setPage(pageFromHash())
        window.addEventListener('hashchange', handleHashChange)
        return () => window.removeEventListener('hashchange', handleHashChange)
    }, [])

    function goToPage(destination) {
        window.location.hash = destination
        setPage(destination)
        setMenuOpen(false)
        window.scrollTo(0, 0)
    }

    function onInvitationAccepted(workspaceID) {
        setSelectedWorkspaceId(workspaceID)
        retryWorkspaces()
        goToPage('workspaces')
    }

    function retryWorkspaces() {
        setWorkspacesLoading(true)
        setWorkspaceLoadError('')
        setReloadVersion((value) => value + 1)
    }

    async function handleUpdateWorkspace(id, settings) {
        const updated = await updateWorkspace(id, settings)
        setWorkspaces((current) => current.map((space) => space.id === id ? updated : space))
        return updated
    }

    useEffect(() => {
        const controller = new AbortController()
        async function load() {
            try {
                const list = await fetchWorkspaces({ signal: controller.signal })
                if (!controller.signal.aborted) setWorkspaces(list)
            } catch (loadError) {
                if (!controller.signal.aborted) {
                    setWorkspaceLoadError(loadError.message || 'Could not load your workspaces.')
                }
            } finally {
                if (!controller.signal.aborted) setWorkspacesLoading(false)
            }
        }
        load()
        return () => controller.abort()
    }, [reloadVersion])

    useEffect(() => {
        const dialog = dialogRef.current
        if (!dialog) return
        if (createDialogOpen && !dialog.open) {
            dialog.showModal()
            nameInputRef.current?.focus()
        } else if (!createDialogOpen && dialog.open) {
            dialog.close()
        }
    }, [createDialogOpen])

    function openCreateDialog() {
        setMenuOpen(false)
        setCreateError('')
        setWorkspaceName('')
        setWorkspacePurpose('general')
        setCreateDialogOpen(true)
    }

    async function handleCreateWorkspace(event) {
        event.preventDefault()
        if (creatingWorkspace) return
        const name = workspaceName.trim()
        if (!name || [...name].length > 100) {
            setCreateError('Use a workspace name between 1 and 100 characters.')
            return
        }

        setCreatingWorkspace(true)
        setCreateError('')
        try {
            const created = await createWorkspace({ name, purpose: workspacePurpose })
            // These records come from PostgreSQL via the Go backend, never demo state.
            setWorkspaces((current) => [created, ...current.filter((space) => space.id !== created.id)])
            setSelectedWorkspaceId(created.id)
            setCreateDialogOpen(false)
            setGuideOpen(false)
        } catch (createFailure) {
            setCreateError(createFailure.message || 'Could not create the workspace.')
        } finally {
            setCreatingWorkspace(false)
        }
    }

    async function handleLogout() {
        setLoggingOut(true)
        setError('')
        try {
            const response = await fetch(`${API}/logout`, {
                method: 'POST',
                credentials: 'include',
            })
            if (response.ok) {
                onLogout()
                return
            }
            setError('Could not log out. Please try again.')
        } catch {
            setError('Could not reach the server. Please try again.')
        }
        setLoggingOut(false)
    }

    return (
        <div className="home-page">
            {menuOpen && (
                <button
                    className="dashboard-mobile-scrim"
                    type="button"
                    aria-label="Close navigation"
                    onClick={() => setMenuOpen(false)}
                />
            )}

            <aside className={`dashboard-sidebar ${menuOpen ? 'dashboard-sidebar--open' : ''}`} aria-label="Dashboard sidebar">
                <div className="dashboard-sidebar-top">
                    <a className="dashboard-brand" href="#overview" onClick={() => goToPage('overview')}>
                        <img src="/noxorbit-mark.svg" alt="" />
                        <span className="dashboard-brand-copy">
                            <strong>NOXORBIT</strong>
                            <small>CONNECT · COLLABORATE · ACHIEVE</small>
                        </span>
                    </a>
                    <button className="dashboard-sidebar-close" type="button" onClick={() => setMenuOpen(false)} aria-label="Close navigation">
                        <X size={20} />
                    </button>
                </div>

                <div className="dashboard-workspace-indicator">
                    <p className="dashboard-sidebar-label">CURRENT WORKSPACE</p>
                    <div className="dashboard-workspace-indicator-body">
                        <span className="dashboard-workspace-icon"><Layers3 size={19} /></span>
                        <span className="dashboard-workspace-indicator-text">
                            <strong title={activeWorkspace?.name}>
                                {workspacesLoading && !hasWorkspace ? 'Loading workspaces…'
                                    : activeWorkspace?.name || (workspaceLoadError ? 'Could not load' : 'No workspace yet')}
                            </strong>
                            <small>{activeWorkspace ? `${purposeLabel(activeWorkspace.purpose)} · ${activeWorkspace.role}`
                                : workspaceLoadError ? 'Retry from your dashboard' : 'Your first one is next'}</small>
                        </span>
                    </div>
                </div>

                <nav className="dashboard-nav" aria-label="Primary">
                    <p className="dashboard-sidebar-label">WORKSPACE</p>
                    <a href="#overview" className={`dashboard-nav-item${page === 'overview' ? ' dashboard-nav-item--active' : ''}`}
                        aria-current={page === 'overview' ? 'page' : undefined}
                        onClick={() => goToPage('overview')}>
                        <LayoutDashboard size={19} aria-hidden="true" />
                        <span>Overview</span>
                        {page === 'overview' && <span className="dashboard-nav-active-mark" aria-hidden="true" />}
                    </a>
                    <a href="#workspaces" className={`dashboard-nav-item${page === 'workspaces' ? ' dashboard-nav-item--active' : ''}`}
                        aria-current={page === 'workspaces' ? 'page' : undefined}
                        onClick={() => goToPage('workspaces')}>
                        <Layers3 size={19} aria-hidden="true" />
                        <span>Workspaces</span>
                        {page === 'workspaces' && <span className="dashboard-nav-active-mark" aria-hidden="true" />}
                    </a>
                    <a href="#projects" className={`dashboard-nav-item${page === 'projects' ? ' dashboard-nav-item--active' : ''}`}
                        aria-current={page === 'projects' ? 'page' : undefined}
                        onClick={() => goToPage('projects')}>
                        <FolderKanban size={19} aria-hidden="true" />
                        <span>Projects</span>
                        {page === 'projects' && <span className="dashboard-nav-active-mark" aria-hidden="true" />}
                    </a>
                    <a href="#my-tasks" className={`dashboard-nav-item${page === 'my-tasks' ? ' dashboard-nav-item--active' : ''}`}
                        aria-current={page === 'my-tasks' ? 'page' : undefined}
                        onClick={() => goToPage('my-tasks')}>
                        <ListTodo size={19} aria-hidden="true" />
                        <span>My tasks</span>
                        {page === 'my-tasks' && <span className="dashboard-nav-active-mark" aria-hidden="true" />}
                    </a>
                    <a href="#activity" className={`dashboard-nav-item${page === 'activity' ? ' dashboard-nav-item--active' : ''}`}
                        aria-current={page === 'activity' ? 'page' : undefined}
                        onClick={() => goToPage('activity')}>
                        <CalendarDays size={19} aria-hidden="true" />
                        <span>Activity</span>
                        {page === 'activity' && <span className="dashboard-nav-active-mark" aria-hidden="true" />}
                    </a>
                    <div className="dashboard-nav-section-separator" aria-hidden="true" />
                    <p className="dashboard-sidebar-label dashboard-manage-label">MANAGE</p>
                    <a href="#recycle-bin" className={`dashboard-nav-item${page === 'recycle-bin' ? ' dashboard-nav-item--active' : ''}`}
                        aria-current={page === 'recycle-bin' ? 'page' : undefined}
                        onClick={() => goToPage('recycle-bin')}>
                        <Trash2 size={19} aria-hidden="true" />
                        <span>Recycle bin</span>
                        {page === 'recycle-bin' && <span className="dashboard-nav-active-mark" aria-hidden="true" />}
                    </a>
                    <a href="#settings" className={`dashboard-nav-item${page === 'settings' ? ' dashboard-nav-item--active' : ''}`}
                        aria-current={page === 'settings' ? 'page' : undefined}
                        onClick={() => goToPage('settings')}>
                        <Settings2 size={19} aria-hidden="true" />
                        <span>Settings</span>
                        {page === 'settings' && <span className="dashboard-nav-active-mark" aria-hidden="true" />}
                    </a>
                </nav>

                <div className="dashboard-sidebar-bottom">
                    <div className="dashboard-sidebar-note">
                        <Sparkles size={19} aria-hidden="true" />
                        <p>Great things start with a shared space.</p>
                    </div>
                    <div className="dashboard-profile">
                        <span className="dashboard-avatar" aria-hidden="true">{initialLetters(user.name)}</span>
                        <div className="dashboard-profile-info">
                            <strong>{user.name}</strong>
                            <span>{user.email}</span>
                        </div>
                    </div>
                    {error && <p className="dashboard-error" role="alert">{error}</p>}
                    <button type="button" className="dashboard-signout" onClick={handleLogout} disabled={loggingOut}>
                        <LogOut size={18} aria-hidden="true" />
                        {loggingOut ? 'Logging out…' : 'Log out'}
                    </button>
                </div>
            </aside>

            <div className="dashboard-area">
                <header className="dashboard-topbar">
                    <div className="dashboard-breadcrumbs">
                        <button className="dashboard-menu-toggle" type="button" onClick={() => setMenuOpen(true)} aria-label="Open navigation" aria-expanded={menuOpen}>
                            <Menu size={22} />
                        </button>
                        <span>{activeWorkspace ? activeWorkspace.name : 'YOUR SPACE'}</span>
                        <ChevronRight size={16} aria-hidden="true" />
                        <strong>{page === 'workspaces' ? 'Workspaces' : page === 'projects' ? 'Projects' : page === 'invite' ? 'Invitation' : page === 'my-tasks' ? 'My tasks' : page === 'recycle-bin' ? 'Recycle bin' : page === 'activity' ? 'Activity' : page === 'settings' ? 'Settings' : 'Overview'}</strong>
                    </div>
                    <div className="dashboard-topbar-account">
                        <span>Signed in</span>
                        <span className="dashboard-avatar dashboard-avatar--small" title={user.name} aria-label={user.name}>{initialLetters(user.name)}</span>
                    </div>
                </header>

                <main className="dashboard-main" id={page === 'overview' ? 'dashboard-overview' : page === 'workspaces' ? 'dashboard-workspaces' : page === 'my-tasks' ? 'dashboard-my-tasks' : page === 'recycle-bin' ? 'dashboard-recycle-bin' : page === 'activity' ? 'dashboard-activity' : page === 'settings' ? 'dashboard-settings' : 'dashboard-projects'}>
                    {page === 'invite' ? (
                        <InvitationPage token={decodeURIComponent(window.location.hash.slice('#invite/'.length))}
                            onAccepted={onInvitationAccepted} />
                    ) : page === 'workspaces' ? (
                        <WorkspacesPage
                            workspaces={workspaces}
                            activeWorkspace={activeWorkspace}
                            loading={workspacesLoading}
                            loadError={workspaceLoadError}
                            onCreate={openCreateDialog}
                            onRetry={retryWorkspaces}
                            onSelect={setSelectedWorkspaceId}
                            onUpdate={handleUpdateWorkspace}
                            onBackToOverview={() => goToPage('overview')}
                        />
                    ) : page === 'my-tasks' ? (
                        <MyTasksPage />
                    ) : page === 'recycle-bin' ? (
                        <RecycleBinPage />
                    ) : page === 'activity' ? (
                        <ActivityPage activeWorkspace={activeWorkspace}
                            workspacesLoading={workspacesLoading}
                            onOpenWorkspaces={() => goToPage('workspaces')} />
                    ) : page === 'settings' ? (
                        <SettingsPage user={user} onAccountDeleted={() => {
                            window.location.hash = ''
                            onLogout()
                        }} />
                    ) : page === 'projects' ? (
                        <ProjectsPage
                            user={user}
                            activeWorkspace={activeWorkspace}
                            workspacesLoading={workspacesLoading}
                            onOpenWorkspaces={() => goToPage('workspaces')}
                        />
                    ) : (
                        <>
                    <div className="dashboard-welcome">
                        <div>
                            <p className="dashboard-eyebrow"><span className="dashboard-live-dot" /> {hasWorkspace ? 'YOUR TEAM SPACE' : 'THE BEGINNING OF SOMETHING GREAT'}</p>
                            <h1>{hasWorkspace ? 'Welcome back' : 'Welcome to your space'}, <span>{displayName}.</span></h1>
                            <p>{hasWorkspace ? `You're viewing ${activeWorkspace.name}. Everything your team builds will start here.` : 'A clearer way to bring your people, projects, and ideas together.'}</p>
                        </div>
                        {!hasWorkspace && !workspacesLoading && !workspaceLoadError && (
                            <span className="dashboard-phase-tag">GETTING STARTED <span> / 01</span></span>
                        )}
                    </div>

                    {workspaceLoadError && (
                        <div className="dashboard-notice" role="alert">
                            <span>{workspaceLoadError}</span>
                            <button type="button" onClick={retryWorkspaces}>Try again</button>
                        </div>
                    )}

                    {hasWorkspace && (
                        <section className="dashboard-active-summary" aria-label="Current workspace">
                            <span className="dashboard-active-summary-icon"><Layers3 size={25} aria-hidden="true" /></span>
                            <div className="dashboard-active-summary-details">
                                <p className="dashboard-eyebrow">CURRENT WORKSPACE</p>
                                <h2>{activeWorkspace.name}</h2>
                                <p>{purposeLabel(activeWorkspace.purpose)} · {activeWorkspace.role === 'owner' ? 'Owner' : 'Member'}</p>
                            </div>
                            <button type="button" className="dashboard-active-summary-action" onClick={() => goToPage('workspaces')}>
                                Manage workspaces <ArrowRight size={17} aria-hidden="true" />
                            </button>
                        </section>
                    )}

                    {!hasWorkspace && (
                    <section className="dashboard-feature" aria-labelledby="dashboard-feature-title">
                        <div className="dashboard-feature-copy">
                            <p className="dashboard-feature-overline"><Compass size={16} aria-hidden="true" /> {hasWorkspace ? 'YOUR WORKSPACE IS READY' : 'YOUR WORK BEGINS HERE'}</p>
                            <h2 id="dashboard-feature-title">{hasWorkspace ? <>A shared space.<br /><em>Now build together.</em></> : <>One place for<br /><em>what comes next.</em></>}</h2>
                            <p>{hasWorkspace
                                ? `${activeWorkspace.name} is ready for your team. Your projects and task boards are ready to use.`
                                : "Create a shared home for your team's work. Projects, tasks, and deadlines will be easier to follow when everything is connected."}</p>
                            <button type="button" className="dashboard-feature-action" onClick={() => goToPage('workspaces')}
                                disabled={workspacesLoading || Boolean(workspaceLoadError)}>
                                {workspacesLoading ? 'Loading your workspaces…' : hasWorkspace ? 'Create another workspace' : 'Create your first workspace'}
                                <ArrowRight size={18} aria-hidden="true" />
                            </button>
                            <button type="button" className="dashboard-feature-secondary" onClick={() => setGuideOpen((open) => !open)}
                                aria-expanded={guideOpen} aria-controls="dashboard-guide">
                                {guideOpen ? 'Hide getting started guide' : 'See how NoxOrbit works'}
                            </button>
                            <span className="dashboard-feature-caption">Your workspaces are saved securely through the NoxOrbit backend.</span>
                        </div>
                        {/* A decorative map of collaboration, not fabricated workspace data.
                            No repeated logo or empty concentric circles. */}
                        <div className="dashboard-collab-art" aria-hidden="true">
                            <span className="dashboard-collab-aura" />
                            <svg className="dashboard-collab-paths" viewBox="0 0 430 330" fill="none" focusable="false">
                                <defs>
                                    <linearGradient id="collab-path-light" x1="30" y1="40" x2="410" y2="310" gradientUnits="userSpaceOnUse">
                                        <stop stopColor="#f4d59c" stopOpacity=".88" />
                                        <stop offset=".5" stopColor="#c983ce" stopOpacity=".65" />
                                        <stop offset="1" stopColor="#ffdaa1" stopOpacity=".9" />
                                    </linearGradient>
                                </defs>
                                <path d="M 103 74 C 116 105 138 109 174 128" stroke="url(#collab-path-light)" strokeWidth="1.5" />
                                <path d="M 326 69 C 313 98 294 100 271 121" stroke="url(#collab-path-light)" strokeWidth="1.5" />
                                <path d="M 276 226 C 300 251 327 252 343 263" stroke="url(#collab-path-light)" strokeWidth="1.5" />
                                <circle cx="174" cy="128" r="3" fill="#f3cf93" />
                                <circle cx="271" cy="121" r="3" fill="#f3cf93" />
                                <circle cx="276" cy="226" r="3" fill="#f3cf93" />
                            </svg>
                            <div className="dashboard-collab-center">
                                <div className="dashboard-collab-center-top">
                                    <span>THE NOXORBIT WAY</span>
                                    <Sparkles size={17} />
                                </div>
                                <div className="dashboard-collab-center-middle">
                                    <span className="dashboard-collab-center-icon"><Layers3 size={26} /></span>
                                    <strong>One shared<br />direction.</strong>
                                </div>
                                <p>A place for everything that matters.</p>
                            </div>
                            <div className="dashboard-collab-chip dashboard-collab-chip--ideas">
                                <span className="dashboard-collab-chip-icon"><Lightbulb size={18} /></span>
                                <span><strong>Ideas</strong><small>Take shape</small></span>
                            </div>
                            <div className="dashboard-collab-chip dashboard-collab-chip--people">
                                <span className="dashboard-collab-chip-icon"><Users size={18} /></span>
                                <span><strong>People</strong><small>Connect</small></span>
                            </div>
                            <div className="dashboard-collab-chip dashboard-collab-chip--progress">
                                <span className="dashboard-collab-chip-icon"><CheckCheck size={18} /></span>
                                <span><strong>Progress</strong><small>Moves forward</small></span>
                            </div>
                            <span className="dashboard-collab-spark dashboard-collab-spark--one" />
                            <span className="dashboard-collab-spark dashboard-collab-spark--two" />
                        </div>
                    </section>
                    )}

                    {!hasWorkspace && guideOpen && (
                        <section className="dashboard-guide" id="dashboard-guide" aria-labelledby="dashboard-guide-title">
                            <div className="dashboard-guide-head">
                                <div>
                                    <p className="dashboard-eyebrow">YOUR FIRST THREE STEPS</p>
                                    <h2 id="dashboard-guide-title">A simple start, a stronger team.</h2>
                                </div>
                                <button type="button" onClick={() => setGuideOpen(false)} aria-label="Close getting started guide">
                                    <X size={20} />
                                </button>
                            </div>
                            <div className="dashboard-guide-steps">
                                {SETUP_STEPS.map((step) => (
                                    <div className="dashboard-guide-step" key={step.number}>
                                        <span>{step.number}</span>
                                        <h3>{step.title}</h3>
                                        <p>{step.description}</p>
                                    </div>
                                ))}
                            </div>
                        </section>
                    )}

                    <div className="dashboard-section-heading">
                        <div>
                            <p className="dashboard-eyebrow">YOUR WORKSPACE AT A GLANCE</p>
                            <h2>{hasWorkspace ? 'Ready for what comes next.' : 'Everything in its place.'}</h2>
                        </div>
                        <span>{hasWorkspace ? activeWorkspace.name : 'Ready when you are'}</span>
                    </div>

                    <div className="dashboard-overview-grid">
                        <ProjectOverviewCard workspaceID={activeWorkspace?.id}
                            onOpen={() => goToPage('projects')} />

                        <OverviewTaskCards onOpenTasks={() => goToPage('my-tasks')} />
                    </div>

                        </>
                    )}
                </main>
            </div>

            <dialog
                ref={dialogRef}
                className="dashboard-workspace-dialog"
                aria-labelledby="dashboard-workspace-dialog-title"
                onClose={() => setCreateDialogOpen(false)}
                onCancel={(event) => { if (creatingWorkspace) event.preventDefault() }}
                onClick={(event) => {
                    if (event.target === event.currentTarget && !creatingWorkspace) setCreateDialogOpen(false)
                }}
            >
                <div className="dashboard-dialog-header">
                    <span className="dashboard-dialog-icon"><Layers3 size={23} aria-hidden="true" /></span>
                    <button type="button" className="dashboard-dialog-close" aria-label="Close create workspace"
                        onClick={() => setCreateDialogOpen(false)} disabled={creatingWorkspace}>
                        <X size={20} aria-hidden="true" />
                    </button>
                </div>
                <p className="dashboard-eyebrow">YOUR SPACE STARTS HERE</p>
                <h2 id="dashboard-workspace-dialog-title">Create a workspace.</h2>
                <p className="dashboard-dialog-description">Give your team a home. You can create more workspaces whenever you need them.</p>
                <form onSubmit={handleCreateWorkspace} className="dashboard-workspace-form">
                    <label htmlFor="dashboard-workspace-name">Workspace name</label>
                    <input
                        ref={nameInputRef}
                        id="dashboard-workspace-name"
                        type="text"
                        name="name"
                        autoComplete="off"
                        value={workspaceName}
                        onChange={(event) => setWorkspaceName(event.target.value)}
                        placeholder="e.g. Study Group A"
                        maxLength={100}
                        required
                        disabled={creatingWorkspace}
                    />
                    <label htmlFor="dashboard-workspace-purpose">What will you use it for?</label>
                    <select
                        id="dashboard-workspace-purpose"
                        value={workspacePurpose}
                        onChange={(event) => setWorkspacePurpose(event.target.value)}
                        disabled={creatingWorkspace}
                    >
                        {WORKSPACE_PURPOSES.map((choice) => (
                            <option key={choice.value} value={choice.value}>{choice.label}</option>
                        ))}
                    </select>
                    <p className="dashboard-form-hint">{WORKSPACE_PURPOSES.find((choice) => choice.value === workspacePurpose)?.detail}</p>
                    {createError && <p className="dashboard-form-error" role="alert">{createError}</p>}
                    <div className="dashboard-dialog-actions">
                        <button type="button" className="dashboard-dialog-cancel" disabled={creatingWorkspace}
                            onClick={() => setCreateDialogOpen(false)}>Cancel</button>
                        <button type="submit" className="dashboard-feature-action dashboard-dialog-submit"
                            disabled={creatingWorkspace || !workspaceName.trim()}>
                            {creatingWorkspace ? <><LoaderCircle size={17} className="dashboard-spin" aria-hidden="true" /> Creating…</>
                                : <><Plus size={17} aria-hidden="true" /> Create workspace</>}
                        </button>
                    </div>
                </form>
            </dialog>
        </div>
    )
}
