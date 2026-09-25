import { useState } from 'react'
import './App.css'

function App() {
  const [form, setForm] = useState({name: "", email: "", password: "", confirmPassword: ""})

  function handleChange(e){
    setForm({...form, [e.target.name]: e.target.value});
  }

  return (
    <>
      <h1>Register for Noxorbit</h1>
      <form onSubmit={handleSubmit}>
      <input type='text' name='name' value={form.name} onChange={handleChange}/>
      <input type='email' name='email' value={form.email} onChange={handleChange} />
      <input type='password' name='password' value={form.password} onChange={handleChange} />
      <input type='password' name='confirmPassword' value={form.confirmPassword} onChange={handleChange} />
      <button type='submit'>Register</button>
      </form>
    </>
  )
}

export default App
