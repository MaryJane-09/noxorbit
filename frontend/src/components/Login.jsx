import { useState } from 'react'
import { Mail, Lock, Eye, EyeOff, ArrowRight } from 'lucide-react'
import AuthShell from './AuthShell'
import { API } from '../api'

export default function Login({ onSuccess, onRegister }) {
    const [form, setForm] = useState({ email: "", password: "" })
    const [message, setMessage] = useState("")
    const [loading, setLoading] = useState(false)
    const [showPassword, setShowPassword] = useState(false)

    function handleChange(e) {
        setForm({ ...form, [e.target.name]: e.target.value })
    }

    async function handleSubmit(e) {
        e.preventDefault()
        setLoading(true)
        setMessage("")
        try {
            const response = await fetch(`${API}/login`, {
                method: "POST",
                credentials: "include",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(form),
            })
            const data = await response.json()
            if (response.ok) {
                await onSuccess()
            } else {
                setMessage(data.error || "Login failed")
            }
        } catch {
            setMessage("Failed to connect to the server.")
        } finally {
            setLoading(false)
        }
    }

    return (
        <AuthShell topbarText="New to Noxorbit?" topbarLabel="Sign up" onTopbar={onRegister}>
            <p className="card-eyebrow">WELCOME BACK</p>
            <h2 className="card-title">Log in to <span className="accent">Noxorbit</span></h2>
            <p className="card-subtitle">Pick up where you left off.</p>

            {message && <p className="form-message" role="alert">{message}</p>}

            <form className="register-form" onSubmit={handleSubmit}>
                <label className="field">
                    <Mail size={18} />
                    <input type="email" name="email" placeholder="Email Address" aria-label="Email address"
                        autoComplete="email" value={form.email} onChange={handleChange} required />
                </label>

                <label className="field">
                    <Lock size={18} />
                    <input type={showPassword ? "text" : "password"} name="password" placeholder="Password"
                        aria-label="Password" autoComplete="current-password"
                        value={form.password} onChange={handleChange} required />
                    <button type="button" className="field-toggle"
                        onClick={() => setShowPassword(!showPassword)}
                        aria-label={showPassword ? "Hide password" : "Show password"}>
                        {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                    </button>
                </label>

                <button type="submit" className="btn-primary" disabled={loading}>
                    {loading ? "Logging in…" : "Log in"} <ArrowRight size={18} />
                </button>
            </form>
        </AuthShell>
    )
}