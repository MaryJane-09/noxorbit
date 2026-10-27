import { useState } from 'react'
import { LogOut } from 'lucide-react'
import { API } from '../api'
import './Register.css'

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
        <div className="register-page">
            <div className="home-wrap">
                <section className="register-card">
                    <p className="card-eyebrow">LOGGED IN</p>
                    <h2 className="card-title">Welcome, <span className="accent">{user.name}</span></h2>
                    <p className="card-subtitle">{user.email}</p>
                    {error && <p className="form-message" role="alert">{error}</p>}
                    <button type="button" className="btn-primary" onClick={handleLogout} disabled={loading}>
                        Log out <LogOut size={18} />
                    </button>
                </section>
            </div>
        </div>
    )
}