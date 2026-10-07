import { useState } from 'react'
import '../App.css'

export default function Register() {
    const [form, setForm] = useState({ name: "", email: "", password: "", confirmPassword: "" })
    const [message, setMessage] = useState("");
    const [step, setStep] = useState("register");
    const [otpCode, setOtpCode] = useState("");

    function handleChange(e) {
        setForm({ ...form, [e.target.name]: e.target.value });
    }

    async function handleSubmit(e) {
        e.preventDefault();
        if (form.password !== form.confirmPassword) {
            setMessage("Passwords do not match!");
            return;
        }
        try {
            const response = await fetch("http://localhost:8080/register", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ name: form.name, email: form.email, password: form.password })
            })
            const data = await response.json();
            if (response.ok) {
                setMessage("");
                setStep("verify");
            } else {
                setMessage(data.error || "Registration failed");
            }
        } catch (error) {
            setMessage("Failed to connect to the server.");
        }
    }

    async function handleVerify(e) {
        e.preventDefault();
        try {
            const response = await fetch("http://localhost:8080/verify-otp", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ email: form.email, code: otpCode })
            })
            const data = await response.json();
            if (response.ok) {
                setMessage(data.message || "Verification successful!");
                setStep("completed");
            } else {
                setMessage(data.error || "Verification failed");
            }
        } catch (error) {
            setMessage("Failed to connect to the server.");
        }
    }

    return (
        <>
            {message && <p>{message}</p>}
            {step === "register" ? (
                <div>
                    <h1>Register for Noxorbit</h1>
                    <form onSubmit={handleSubmit}>
                        <input type='text' name='name' placeholder='Name' value={form.name} onChange={handleChange} required />
                        <input type='email' name='email' placeholder='Email' value={form.email} onChange={handleChange} required />
                        <input type='password' name='password' placeholder='Password' value={form.password} onChange={handleChange} required />
                        <input type='password' name='confirmPassword' placeholder='Confirm Password' value={form.confirmPassword} onChange={handleChange} required />
                        <button type='submit'>Register</button>
                    </form>
                </div>
            ) : (
                <div>
                    <h1>OTP form here</h1>
                    <form onSubmit={handleVerify}>
                        <input type='text' name='otp' placeholder='Enter OTP Code' value={otpCode} onChange={(e) => setOtpCode(e.target.value)} required />
                        <button type='submit'> Verify OTP</button>
                    </form>
                </div>
            )}
        </>
    )
}