import { useState } from 'react'
import {
    ArrowRight,
    ArrowUpRight,
    CalendarDays,
    ChevronRight,
    Compass,
    FolderKanban,
    LayoutDashboard,
    Layers3,
    ListTodo,
    LogOut,
    Menu,
    Sparkles,
    Users,
    X,
} from 'lucide-react'
import { API } from '../api'
import './Home.css'

// Sections that need workspace data are intentionally informational for now.
// They will become links only after their authenticated APIs exist.
const PLANNED_SECTIONS = [
    { name: 'Workspaces', icon: Layers3 },
    { name: 'Projects', icon: FolderKanban },
    { name: 'My tasks', icon: ListTodo },
    { name: 'Activity', icon: CalendarDays },
]

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

export default function Home({ user, onLogout }) {
    const [loggingOut, setLoggingOut] = useState(false)
    const [error, setError] = useState('')
    const [guideOpen, setGuideOpen] = useState(false)
    const [menuOpen, setMenuOpen] = useState(false)
    const displayName = (user.name || '').trim().split(/\s+/)[0] || 'there'

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
                    <a className="dashboard-brand" href="#dashboard-overview" onClick={() => setMenuOpen(false)}>
                        <img src="/noxorbit-mark.svg" alt="" />
                        <span className="dashboard-brand-copy">
                            <strong>NOXORBIT</strong>
                            <small>YOUR SPACE TO BUILD</small>
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
                        <span>
                            <strong>No workspace yet</strong>
                            <small>Your first one is next</small>
                        </span>
                    </div>
                </div>

                <nav className="dashboard-nav" aria-label="Primary">
                    <p className="dashboard-sidebar-label">WORKSPACE</p>
                    <a href="#dashboard-overview" className="dashboard-nav-item dashboard-nav-item--active" aria-current="page" onClick={() => setMenuOpen(false)}>
                        <LayoutDashboard size={19} />
                        <span>Overview</span>
                        <span className="dashboard-nav-active-mark" aria-hidden="true" />
                    </a>
                    {PLANNED_SECTIONS.map(({ name, icon: Icon }) => (
                        <span className="dashboard-nav-item dashboard-nav-item--planned" key={name} title="Available when workspace features are connected">
                            <Icon size={19} aria-hidden="true" />
                            <span>{name}</span>
                            <small>SOON</small>
                        </span>
                    ))}
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
                        <span>YOUR SPACE</span>
                        <ChevronRight size={16} aria-hidden="true" />
                        <strong>Overview</strong>
                    </div>
                    <div className="dashboard-topbar-account">
                        <span>Signed in</span>
                        <span className="dashboard-avatar dashboard-avatar--small" title={user.name} aria-label={user.name}>{initialLetters(user.name)}</span>
                    </div>
                </header>

                <main className="dashboard-main" id="dashboard-overview">
                    <div className="dashboard-welcome">
                        <div>
                            <p className="dashboard-eyebrow"><span className="dashboard-live-dot" /> THE BEGINNING OF SOMETHING GREAT</p>
                            <h1>Welcome to your space, <span>{displayName}.</span></h1>
                            <p>A clearer way to bring your people, projects, and ideas together.</p>
                        </div>
                        <span className="dashboard-phase-tag">GETTING STARTED <span> / 01</span></span>
                    </div>

                    <section className="dashboard-feature" aria-labelledby="dashboard-feature-title">
                        <div className="dashboard-feature-copy">
                            <p className="dashboard-feature-overline"><Compass size={16} aria-hidden="true" /> YOUR WORK BEGINS HERE</p>
                            <h2 id="dashboard-feature-title">One place for<br /><em>what comes next.</em></h2>
                            <p>Create a shared home for your team's work. Projects, tasks, and deadlines will be easier to follow when everything is connected.</p>
                            <button type="button" className="dashboard-feature-action" onClick={() => setGuideOpen((open) => !open)} aria-expanded={guideOpen} aria-controls="dashboard-guide">
                                {guideOpen ? 'Hide your starting guide' : 'See how to get started'}
                                <ArrowRight size={18} aria-hidden="true" />
                            </button>
                            <span className="dashboard-feature-caption">Workspace creation will be connected in the next development step.</span>
                        </div>
                        <div className="dashboard-orbit" aria-hidden="true">
                            <span className="dashboard-orbit-outer" />
                            <span className="dashboard-orbit-middle" />
                            <span className="dashboard-orbit-inner" />
                            <span className="dashboard-orbit-core"><img src="/noxorbit-mark.svg" alt="" /></span>
                            <span className="dashboard-orbit-star dashboard-orbit-star--one" />
                            <span className="dashboard-orbit-star dashboard-orbit-star--two" />
                        </div>
                    </section>

                    {guideOpen && (
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
                            <h2>Everything in its place.</h2>
                        </div>
                        <span>Ready when you are</span>
                    </div>

                    <div className="dashboard-overview-grid">
                        <section className="dashboard-empty-card" aria-labelledby="dashboard-projects-heading">
                            <div className="dashboard-empty-top">
                                <span className="dashboard-tile-icon"><FolderKanban size={23} aria-hidden="true" /></span>
                                <ArrowUpRight size={19} aria-hidden="true" />
                            </div>
                            <h3 id="dashboard-projects-heading">Projects</h3>
                            <p>Your projects will find their home here once you create a workspace.</p>
                            <span className="dashboard-empty-status">Not set up yet</span>
                        </section>

                        <section className="dashboard-empty-card" aria-labelledby="dashboard-tasks-heading">
                            <div className="dashboard-empty-top">
                                <span className="dashboard-tile-icon"><ListTodo size={23} aria-hidden="true" /></span>
                                <ArrowUpRight size={19} aria-hidden="true" />
                            </div>
                            <h3 id="dashboard-tasks-heading">My tasks</h3>
                            <p>Know what you're responsible for and what deserves your attention.</p>
                            <span className="dashboard-empty-status">Not set up yet</span>
                        </section>

                        <section className="dashboard-empty-card" aria-labelledby="dashboard-deadlines-heading">
                            <div className="dashboard-empty-top">
                                <span className="dashboard-tile-icon"><CalendarDays size={23} aria-hidden="true" /></span>
                                <ArrowUpRight size={19} aria-hidden="true" />
                            </div>
                            <h3 id="dashboard-deadlines-heading">Upcoming deadlines</h3>
                            <p>Your important dates will appear here when tasks are connected.</p>
                            <span className="dashboard-empty-status">Not set up yet</span>
                        </section>
                    </div>

                    <div className="dashboard-footer-hint">
                        <Users size={19} aria-hidden="true" />
                        <span>Better teams build bigger dreams. Your workspace is where yours will take shape.</span>
                    </div>
                </main>
            </div>
        </div>
    )
}
