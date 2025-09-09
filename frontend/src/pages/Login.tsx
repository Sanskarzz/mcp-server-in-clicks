import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../state/auth'

export default function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const nav = useNavigate()
  const loc = useLocation()
  const { login } = useAuth()

  function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    // Demo: accept any credentials
    login()
    const dest = (loc.state as any)?.from || '/dashboard'
    nav(dest)
  }

  return (
    <div className="min-h-screen grid place-items-center gradient-hero">
      <div className="bg-white text-gray-900 rounded-2xl shadow-card w-full max-w-md p-8">
        <h1 className="text-xl font-semibold mb-4">Welcome Back</h1>
        <form onSubmit={onSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Email</label>
            <input type="email" value={email} onChange={e=>setEmail(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Password</label>
            <input type="password" value={password} onChange={e=>setPassword(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
          </div>
          <button className="w-full bg-brand-600 text-white rounded-xl py-2">Sign In</button>
          <button type="button" onClick={()=>{ login(); nav('/dashboard') }} className="w-full bg-gray-100 rounded-xl py-2">Quick Demo Login</button>
          <div className="text-center text-sm text-gray-600">Don't have an account? <Link to="/register" className="text-brand-700">Sign up</Link></div>
        </form>
      </div>
    </div>
  )
}

