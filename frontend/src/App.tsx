import { Suspense, lazy } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { Toaster } from './components/Toaster'
import AppShell from './components/layout/AppShell'
import AuthGuard from './components/auth/AuthGuard'

const Login = lazy(() => import('./pages/Login'))
const Dashboard = lazy(() => import('./pages/Dashboard'))
const Register = lazy(() => import('./pages/Register'))
const NewServer = lazy(() => import('./pages/servers/NewServer'))
const ServerDetail = lazy(() => import('./pages/servers/ServerDetail'))

export default function App() {
  return (
    <Suspense fallback={<div className="p-8">Loading...</div>}>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />

        <Route element={<AuthGuard />}> 
          <Route element={<AppShell />}> 
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/servers/new" element={<NewServer />} />
          <Route path="/servers/:id" element={<ServerDetail />} />
          </Route>
        </Route>

        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
      <Toaster />
    </Suspense>
  )
}

