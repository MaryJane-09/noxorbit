import { Users, SquareCheck, Zap } from 'lucide-react'
import './Register.css'

const FEATURES = [
    { icon: <Users size={20} />, title: 'Work together', text: 'Keep your team in sync' },
    { icon: <SquareCheck size={20} />, title: 'Stay organized', text: 'Manage tasks, projects and deadlines' },
    { icon: <Zap size={20} />, title: 'Turn ideas into action', text: 'Build, create and achieve more' },
]

export default function AuthShell({ topbarText, topbarLabel, onTopbar, children }) {
    return (
        <div className="register-page">
            <div className="register-topbar">
                <span>{topbarText}</span>
                <button type="button" className="btn-outline" onClick={() => onTopbar?.()}>
                    {topbarLabel}
                </button>
            </div>

            <div className="register-layout">
                <section className="register-hero">
                    <h1>
                        Everything you need,<br />
                        <span className="accent">in one place.</span>
                    </h1>
                    <p className="hero-sub">
                        Noxorbit brings people, teams, projects and ideas together — one central workspace.
                    </p>

                    <ul className="feature-list">
                        {FEATURES.map((f) => (
                            <li className="feature" key={f.title}>
                                <span className="feature-icon">{f.icon}</span>
                                <div>
                                    <span className="feature-title">{f.title}</span>
                                    <span className="feature-text">{f.text}</span>
                                </div>
                            </li>
                        ))}
                    </ul>

                    <p className="hero-footer">Better teams. Bigger dreams.</p>
                </section>

                <section className="register-card">{children}</section>
            </div>
        </div>
    )
}