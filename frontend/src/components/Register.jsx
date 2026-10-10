import { useState, useEffect } from 'react'
import {
    User, Mail, Lock, Eye, EyeOff, ArrowRight, ArrowLeft, ShieldCheck, CircleCheck,
    Users, SquareCheck, Zap,
} from 'lucide-react'
import './Register.css'

const OTP_LENGTH = 8
const OTP_LIFETIME = 300 // seconds — keep in sync with the 5-minute expiry in the Go otp package
const RESEND_AFTER = 30  // seconds before "Resend code" unlocks

const FEATURES = [
    { icon: <Users size={20} />, title: 'Work together', text: 'Keep your team in sync' },
    { icon: <SquareCheck size={20} />, title: 'Stay organized', text: 'Manage tasks, projects and deadlines' },
    { icon: <Zap size={20} />, title: 'Turn ideas into action', text: 'Build, create and achieve more' },
]

function formatTime(total) {
    const m = String(Math.floor(total / 60)).padStart(2, "0")
    const s = String(total % 60).padStart(2, "0")
    return `${m}:${s}`
}

export default function Register({ onLogin }) {
    const [form, setForm] = useState({ name: "", email: "", password: "", confirmPassword: "" })
    const [message, setMessage] = useState("")
    const [notice, setNotice] = useState("")
    const [step, setStep] = useState("register")
    const [otpCode, setOtpCode] = useState("")
    const [secondsLeft, setSecondsLeft] = useState(OTP_LIFETIME)
    const [loading, setLoading] = useState(false)
    const [showPassword, setShowPassword] = useState(false)
    const [showConfirm, setShowConfirm] = useState(false)

    // Tick once a second, only while the verify screen is showing.
    useEffect(() => {
        if (step !== "verify") return
        const id = setInterval(() => {
            setSecondsLeft((s) => Math.max(s - 1, 0))
        }, 1000)
        return () => clearInterval(id)
    }, [step])

    // Derived values: computed from state on every render, so they need no state of their own.
    const expired = secondsLeft === 0
    const resendWait = Math.max(RESEND_AFTER - (OTP_LIFETIME - secondsLeft), 0)

    function handleChange(e) {
        setForm({ ...form, [e.target.name]: e.target.value })
    }

    // Used by both the first registration and "Resend code".
    async function sendRegistration() {
        const response = await fetch("http://localhost:8080/register", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name: form.name, email: form.email, password: form.password })
        })
        const data = await response.json()
        return { ok: response.ok, data }
    }

    async function handleSubmit(e) {
        e.preventDefault()
        if (form.password !== form.confirmPassword) {
            setMessage("Passwords do not match!")
            return
        }
        setLoading(true)
        try {
            const { ok, data } = await sendRegistration()
            if (ok) {
                setMessage("")
                setNotice("")
                setOtpCode("")
                setSecondsLeft(OTP_LIFETIME)
                setStep("verify")
            } else {
                setMessage(data.error || "Registration failed")
            }
        } catch {
            setMessage("Failed to connect to the server.")
        } finally {
            setLoading(false)
        }
    }

    async function handleResend() {
        setLoading(true)
        setMessage("")
        setNotice("")
        try {
            const { ok, data } = await sendRegistration()
            if (ok) {
                setOtpCode("")
                setSecondsLeft(OTP_LIFETIME)
                setNotice("A new code is on its way.")
            } else {
                setMessage(data.error || "Could not send a new code")
            }
        } catch {
            setMessage("Failed to connect to the server.")
        } finally {
            setLoading(false)
        }
    }

    async function handleVerify(e) {
        e.preventDefault()
        setLoading(true)
        setNotice("")
        try {
            const response = await fetch("http://localhost:8080/verify-otp", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ email: form.email, code: otpCode.trim() })
            })
            const data = await response.json()
            if (response.ok) {
                setMessage("")
                setStep("completed")
            } else {
                setMessage(data.error || "Verification failed")
            }
        } catch {
            setMessage("Failed to connect to the server.")
        } finally {
            setLoading(false)
        }
    }

    function handleChangeEmail() {
        setMessage("")
        setNotice("")
        setOtpCode("")
        setStep("register")
    }

    return (
        <div className="register-page">
            <div className="register-topbar">
                <span>Already have an account?</span>
                <button type="button" className="btn-outline" onClick={() => onLogin?.()}>Login</button>
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

                <section className="register-card">
                    {step === "register" && (
                        <>
                            <p className="card-eyebrow">CREATE YOUR ACCOUNT</p>
                            <h2 className="card-title">Join <span className="accent">Noxorbit</span></h2>
                            <p className="card-subtitle">Start your journey today.</p>

                            {message && <p className="form-message" role="alert">{message}</p>}

                            <form className="register-form" onSubmit={handleSubmit}>
                                <label className="field">
                                    <User size={18} />
                                    <input type="text" name="name" placeholder="Full Name" aria-label="Full name"
                                        autoComplete="name" value={form.name} onChange={handleChange} required />
                                </label>

                                <label className="field">
                                    <Mail size={18} />
                                    <input type="email" name="email" placeholder="Email Address" aria-label="Email address"
                                        autoComplete="email" value={form.email} onChange={handleChange} required />
                                </label>

                                <label className="field">
                                    <Lock size={18} />
                                    <input type={showPassword ? "text" : "password"} name="password" placeholder="Password"
                                        aria-label="Password" autoComplete="new-password"
                                        value={form.password} onChange={handleChange} required />
                                    <button type="button" className="field-toggle"
                                        onClick={() => setShowPassword(!showPassword)}
                                        aria-label={showPassword ? "Hide password" : "Show password"}>
                                        {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                                    </button>
                                </label>

                                <label className="field">
                                    <Lock size={18} />
                                    <input type={showConfirm ? "text" : "password"} name="confirmPassword" placeholder="Confirm Password"
                                        aria-label="Confirm password" autoComplete="new-password"
                                        value={form.confirmPassword} onChange={handleChange} required />
                                    <button type="button" className="field-toggle"
                                        onClick={() => setShowConfirm(!showConfirm)}
                                        aria-label={showConfirm ? "Hide password" : "Show password"}>
                                        {showConfirm ? <EyeOff size={18} /> : <Eye size={18} />}
                                    </button>
                                </label>

                                <button type="submit" className="btn-primary" disabled={loading}>
                                    Create Account <ArrowRight size={18} />
                                </button>
                            </form>

                            <p className="card-terms">
                                By creating an account, you agree to our{" "}
                                <span className="accent">Terms of Service</span> and{" "}
                                <span className="accent">Privacy Policy</span>.
                            </p>
                        </>
                    )}

                    {step === "verify" && (
                        <>
                            <p className="card-eyebrow">VERIFY YOUR EMAIL</p>
                            <h2 className="card-title">Check your <span className="accent">inbox</span></h2>
                            <p className="card-subtitle">
                                We sent a {OTP_LENGTH}-character code to <strong>{form.email}</strong>.
                            </p>

                            {message && <p className="form-message" role="alert">{message}</p>}
                            {notice && <p className="form-notice" role="status">{notice}</p>}

                            <form className="register-form" onSubmit={handleVerify}>
                                <input
                                    className="otp-input"
                                    type="text"
                                    name="otp"
                                    placeholder="········"
                                    aria-label="Verification code"
                                    autoComplete="one-time-code"
                                    autoCapitalize="off"
                                    autoCorrect="off"
                                    spellCheck={false}
                                    maxLength={OTP_LENGTH}
                                    value={otpCode}
                                    onChange={(e) => setOtpCode(e.target.value)}
                                    required
                                />

                                <p className={`otp-timer ${expired ? "is-expired" : ""}`}>
                                    {expired
                                        ? "This code has expired. Request a new one."
                                        : `Code expires in ${formatTime(secondsLeft)}`}
                                </p>

                                <button
                                    type="submit"
                                    className="btn-primary"
                                    disabled={loading || expired || otpCode.trim().length !== OTP_LENGTH}
                                >
                                    Verify <ArrowRight size={18} />
                                </button>
                            </form>

                            <div className="otp-actions">
                                <button type="button" className="link-btn" onClick={handleChangeEmail}>
                                    <ArrowLeft size={16} /> Change email
                                </button>
                                <button type="button" className="link-btn" onClick={handleResend}
                                    disabled={loading || resendWait > 0}>
                                    {resendWait > 0 ? `Resend in ${resendWait}s` : "Resend code"}
                                </button>
                            </div>
                        </>
                    )}

                    {step === "completed" && (
                        <div className="success-state">
                            <span className="success-icon"><CircleCheck size={48} /></span>
                            <p className="card-eyebrow">EMAIL VERIFIED</p>
                            <h2 className="card-title">Welcome to <span className="accent">Noxorbit</span></h2>
                            <p className="card-subtitle">Your account is ready.</p>
                            <button type="button" className="btn-primary" onClick={() => onLogin?.()}>
                                Continue to login <ArrowRight size={18} />
                            </button>
                        </div>
                    )}
                </section>
            </div>
        </div>
    )
}