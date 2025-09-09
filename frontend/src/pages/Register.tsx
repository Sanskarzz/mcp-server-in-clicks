import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../state/auth'

export default function Register() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const nav = useNavigate()
  const { login } = useAuth()

  function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (password !== confirm) return (window as any).__pushToast?.('Passwords do not match')
    // Demo only: mark as authenticated
    login()
    nav('/dashboard')
  }

  return (
    <div className="min-h-screen grid place-items-center gradient-hero">
      <div className="bg-white text-gray-900 rounded-2xl shadow-card w-full max-w-md p-8">
        <h1 className="text-xl font-semibold mb-4">Create your account</h1>
        <form onSubmit={onSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Email</label>
            <input type="email" value={email} onChange={e=>setEmail(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Password</label>
            <input type="password" value={password} onChange={e=>setPassword(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Confirm Password</label>
            <input type="password" value={confirm} onChange={e=>setConfirm(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
          </div>
          <button className="w-full bg-brand-600 text-white rounded-xl py-2">Sign Up</button>
          <button type="button" onClick={()=>nav('/login')} className="w-full bg-gray-100 rounded-xl py-2">Back to Sign In</button>
        </form>
      </div>
    </div>
  )
}

