import { Users, SquareCheck, ChartNoAxesColumnIncreasing } from 'lucide-react'
import './Register.css'

const FEATURES = [
    { icon: <Users size={20} />, title: 'Work together', text: 'Keep your team in sync' },
    { icon: <SquareCheck size={20} />, title: 'Stay organized', text: 'Manage tasks, projects and deadlines' },
    { icon: <ChartNoAxesColumnIncreasing size={20} />, title: 'Achieve more', text: 'Turn ideas into real progress' },
]

export default function AuthShell({ topbarText, topbarLabel, onTopbar, children }) {
    return (
        <div className="register-page auth-page">
            <div className="register-topbar">
                <span>{topbarText}</span>
                <button type="button" className="btn-outline" onClick={() => onTopbar?.()}>
                    {topbarLabel}
                </button>
            </div>

            <div className="register-layout">
                <section className="register-hero">
                    <h1>
                        Everything you<br />
                        need,<br />
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
                </section>

                <section className="register-card">{children}</section>
            </div>
        </div>
    )
}