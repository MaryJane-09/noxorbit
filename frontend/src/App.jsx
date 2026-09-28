import { useState } from 'react'
import './App.css'

function App() {
  const [form, setForm] = useState({ name: "", email: "", password: "", confirmPassword: "" })

  function handleChange(e) {
    setForm({ ...form, [e.target.name]: e.target.value });
  }

  async function handleSubmit(e){
    e.preventDefault();

    const response = await fetch("http://localhost:8080/register", {
      method : "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({name: form.name, email: form.email, password: form.password})
    })

    const data = await response.json();
    console.log(data)

  }

  return (
    <>
      <h1>Register for Noxorbit</h1>
      <form onSubmit={handleSubmit}>
        <input type='text' name='name' value={form.name} onChange={handleChange} />
        <input type='email' name='email' value={form.email} onChange={handleChange} />
        <input type='password' name='password' value={form.password} onChange={handleChange} />
        <input type='password' name='confirmPassword' value={form.confirmPassword} onChange={handleChange} />
        <button type='submit'>Register</button>
      </form>
    </>
  )
}

export default App