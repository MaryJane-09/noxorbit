import { useState } from 'react'
import { LogOut } from 'lucide-react'
import { API } from '../api'
import AuthCosmos from './AuthCosmos'
import './Register.css'
import './Home.css'

export default function Home({ user, onLogout }) {
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState("")

    async function handleLogout() {
        setLoading(true)
        setError("")
        try {
            const response = await fetch(`${API}/logout`, { method: "POST", credentials: "include" })
            if (response.ok) {
                onLogout()
                return
            }
            setError("Could not log out. Try again.")
        } catch {
            setError("Failed to connect to the server.")
        }
        setLoading(false)
    }

    return (
        <div className="register-page home-page">
            <AuthCosmos />
            <main className="home-wrap">
                <section className="register-card home-card" aria-labelledby="home-welcome">
                    <span className="card-edge-light card-edge-light--top" aria-hidden="true" />
                    <span className="card-edge-light card-edge-light--bottom" aria-hidden="true" />
                    <span className="card-edge-light card-edge-light--corner" aria-hidden="true" />
                    <p className="card-eyebrow">LOGGED IN</p>
                    <h1 id="home-welcome" className="card-title">Welcome, <span className="accent">{user.name}</span></h1>
                    <p className="card-subtitle home-email">{user.email}</p>
                    {error && <p className="form-message" role="alert">{error}</p>}
                    <div className="home-actions">
                        <button type="button" className="btn-primary" onClick={handleLogout} disabled={loading}>
                            {loading ? "Logging out…" : "Log out"} <LogOut size={18} />
                        </button>
                    </div>
                </section>
            </main>
        </div>
    )
}