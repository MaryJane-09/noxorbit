import { useState, useEffect } from 'react'
import { Mail, Lock, Eye, EyeOff, ShieldCheck, ArrowRight, CircleCheck } from 'lucide-react'
import AuthShell from './AuthShell'
import { API } from '../api'

const RESEND_AFTER = 60 
const CODE_LENGTH = 6

export default function ForgotPassword({ onLogin }) {
    const [form, setForm] = useState({ email: "", code: "", newPassword: "", confirmPassword: "" })
    const [sending, setSending] = useState(false)
    const [sent, setSent] = useState(false)
    const [cooldown, setCooldown] = useState(0)
    const [message, setMessage] = useState("")
    const [notice, setNotice] = useState("")
    const [loading, setLoading] = useState(false)
    const [done, setDone] = useState(false)
    const [showPassword, setShowPassword] = useState(false)

    useEffect(() => {
        if (cooldown <= 0) return
        const id = setTimeout(() => setCooldown((c) => c - 1), 1000)
        return () => clearTimeout(id)
    }, [cooldown])

    function handleChange(e) {
        setForm({ ...form, [e.target.name]: e.target.value })
    }

    async function handleSend() {
        if (!form.email.includes("@")) {
            setMessage("Enter your email address first.")
            return
        }
        setSending(true)
        setMessage("")
        try {
            const response = await fetch(`${API}/forgot-password`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ email: form.email }),
            })
            const data = await response.json()
            if (response.ok) {
                setSent(true)
                setCooldown(RESEND_AFTER)
                setNotice(data.message)
            } else {
                setMessage(data.error || "Could not send a code")
            }
        } catch {
            setMessage("Failed to connect to the server.")
        } finally {
            setSending(false)
        }
    }

    function handleChangeEmail() {
        setSent(false)
        setCooldown(0)
        setNotice("")
        setMessage("")
        setForm({ ...form, code: "" })
    }

    async function handleSubmit(e) {
        e.preventDefault()
        if (form.newPassword !== form.confirmPassword) {
            setMessage("Passwords do not match!")
            return
        }
        setLoading(true)
        setMessage("")
        try {
            const response = await fetch(`${API}/reset-password`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ email: form.email, code: form.code.trim(), newPassword: form.newPassword }),
            })
            const data = await response.json()
            if (response.ok) {
                setDone(true)
            } else {
                setMessage(data.error || "Could not reset password")
            }
        } catch {
            setMessage("Failed to connect to the server.")
        } finally {
            setLoading(false)
        }
    }

    let sendLabel = "Send code"
    if (sending) sendLabel = "Sending…"
    else if (sent && cooldown > 0) sendLabel = `Sent ✓ · resend in ${cooldown}s`
    else if (sent) sendLabel = "Resend code"

    return (
        <AuthShell topbarText="Remembered it?" topbarLabel="Log in" onTopbar={onLogin}>
            {done ? (
                <div className="success-state">
                    <span className="success-icon"><CircleCheck size={48} /></span>
                    <p className="card-eyebrow">PASSWORD UPDATED</p>
                    <h2 className="card-title">You are <span className="accent">all set</span></h2>
                    <p className="card-subtitle">Log in with your new password. You have been signed out everywhere else.</p>
                    <button type="button" className="btn-primary" onClick={onLogin}>
                        Back to login <ArrowRight size={18} />
                    </button>
                </div>
            ) : (
                <>
                    <p className="card-eyebrow">RESET YOUR PASSWORD</p>
                    <h2 className="card-title">Forgot your <span className="accent">password?</span></h2>
                    <p className="card-subtitle">We will email you a one-time code.</p>

                    {message && <p className="form-message" role="alert">{message}</p>}
                    {notice && <p className="form-notice" role="status">{notice}</p>}

                    <form className="register-form" onSubmit={handleSubmit}>
                        <label className="field">
                            <Mail size={18} />
                            <input type="email" name="email" placeholder="Email Address" aria-label="Email address"
                                autoComplete="email" value={form.email} onChange={handleChange}
                                disabled={sent} required />
                        </label>

                        <button type="button"
                            className={`btn-secondary ${sent && cooldown > 0 ? "is-sent" : ""}`}
                            onClick={handleSend}
                            disabled={sending || cooldown > 0 || !form.email}>
                            {sendLabel}
                        </button>

                        {sent && (
                            <button type="button" className="link-btn" onClick={handleChangeEmail}>
                                Use a different email
                            </button>
                        )}

                        <label className={`field ${!sent ? "is-disabled" : ""}`}>
                            <ShieldCheck size={18} />
                            <input type="text" name="code" aria-label="Verification code"
                                placeholder={sent ? "Enter the code" : "Send a code first"}
                                autoComplete="one-time-code" autoCapitalize="off" autoCorrect="off" spellCheck={false}
                                maxLength={CODE_LENGTH} value={form.code} onChange={handleChange}
                                disabled={!sent} required />
                        </label>

                        <label className={`field ${!sent ? "is-disabled" : ""}`}>
                            <Lock size={18} />
                            <input type={showPassword ? "text" : "password"} name="newPassword"
                                placeholder="New password" aria-label="New password" autoComplete="new-password"
                                value={form.newPassword} onChange={handleChange} disabled={!sent} required />
                            <button type="button" className="field-toggle" disabled={!sent}
                                onClick={() => setShowPassword(!showPassword)}
                                aria-label={showPassword ? "Hide password" : "Show password"}>
                                {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                            </button>
                        </label>

                        <label className={`field ${!sent ? "is-disabled" : ""}`}>
                            <Lock size={18} />
                            <input type={showPassword ? "text" : "password"} name="confirmPassword"
                                placeholder="Confirm new password" aria-label="Confirm new password"
                                autoComplete="new-password" value={form.confirmPassword}
                                onChange={handleChange} disabled={!sent} required />
                        </label>

                        <button type="submit" className="btn-primary"
                            disabled={!sent || loading || form.code.trim().length !== CODE_LENGTH}>
                            {loading ? "Updating…" : "Reset password"} <ArrowRight size={18} />
                        </button>
                    </form>
                </>
            )}
        </AuthShell>
    )
}