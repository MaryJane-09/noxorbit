import { useState } from 'react'
import {
    User, Mail, Lock, Eye, EyeOff, ArrowRight, ShieldCheck,
    Users, SquareCheck, Zap,
} from 'lucide-react'
import './Register.css'

const FEATURES = [
    { icon: <Users size={20} />, title: 'Work together', text: 'Keep your team in sync' },
    { icon: <SquareCheck size={20} />, title: 'Stay organized', text: 'Manage tasks, projects and deadlines' },
    { icon: <Zap size={20} />, title: 'Turn ideas into action', text: 'Build, create and achieve more' },
]

export default function Register() {
    const [form, setForm] = useState({ name: "", email: "", password: "", confirmPassword: "" })
    const [message, setMessage] = useState("")
    const [step, setStep] = useState("register")
    const [otpCode, setOtpCode] = useState("")
    const [showPassword, setShowPassword] = useState(false)
    const [showConfirm, setShowConfirm] = useState(false)

    function handleChange(e) {
        setForm({ ...form, [e.target.name]: e.target.value })
    }

    async function handleSubmit(e) {
        e.preventDefault()
        if (form.password !== form.confirmPassword) {
            setMessage("Passwords do not match!")
            return
        }
        try {
            const response = await fetch("http://localhost:8080/register", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ name: form.name, email: form.email, password: form.password })
            })
            const data = await response.json()
            if (response.ok) {
                setMessage("")
                setStep("verify")
            } else {
                setMessage(data.error || "Registration failed")
            }
        } catch (error) {
            setMessage("Failed to connect to the server.")
        }
    }

    async function handleVerify(e) {
        e.preventDefault()
        try {
            const response = await fetch("http://localhost:8080/verify-otp", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ email: form.email, code: otpCode })
            })
            const data = await response.json()
            if (response.ok) {
                setMessage(data.message || "Verification successful!")
                setStep("completed")
            } else {
                setMessage(data.error || "Verification failed")
            }
        } catch (error) {
            setMessage("Failed to connect to the server.")
        }
    }

    return (
        <div className="register-page">
            <div className="register-topbar">
                <span>Already have an account?</span>
                <button type="button" className="btn-outline">Login</button>
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

                                <button type="submit" className="btn-primary">
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
                            <p className="card-subtitle">We sent a code to {form.email}.</p>

                            {message && <p className="form-message" role="alert">{message}</p>}

                            <form className="register-form" onSubmit={handleVerify}>
                                <label className="field">
                                    <ShieldCheck size={18} />
                                    <input type="text" placeholder="Enter OTP code" aria-label="OTP code"
                                        autoComplete="one-time-code" spellCheck={false}
                                        value={otpCode} onChange={(e) => setOtpCode(e.target.value)} required />
                                </label>
                                <button type="submit" className="btn-primary">
                                    Verify <ArrowRight size={18} />
                                </button>
                            </form>
                        </>
                    )}

                    {step === "completed" && (
                        <>
                            <p className="card-eyebrow">ALL SET</p>
                            <h2 className="card-title">Welcome to <span className="accent">Noxorbit</span></h2>
                            <p className="card-subtitle">{message}</p>
                        </>
                    )}
                </section>
            </div>
        </div>
    )
}