import { Outlet, Link, useLocation } from 'react-router-dom'
import { Server, PlusCircle } from 'lucide-react'

export default function AppShell() {
  const loc = useLocation()
  return (
    <div className="min-h-screen grid grid-cols-[240px_1fr] grid-rows-[56px_1fr]">
      <header className="col-span-2 h-14 flex items-center justify-between px-6 gradient-hero text-white shadow">
        <Link to="/dashboard" className="font-semibold">MCP SaaS</Link>
        <div className="flex items-center gap-3">
          <Link to="/servers/new" className="inline-flex items-center gap-1 bg-white/10 hover:bg-white/20 px-3 py-1.5 rounded-xl">
            <PlusCircle size={18}/> New Server
          </Link>
        </div>
      </header>
      <aside className="border-r bg-white/70 backdrop-blur p-4">
        <nav className="space-y-1 text-sm">
          <NavItem to="/dashboard" label="Dashboard" active={loc.pathname.startsWith('/dashboard')} />
          <NavItem to="/servers/new" label="Create Server" active={loc.pathname.startsWith('/servers/new')} />
        </nav>
      </aside>
      <main className="p-6 bg-gray-50">
        <Outlet />
      </main>
    </div>
  )
}

function NavItem({ to, label, active }: { to: string, label: string, active: boolean }) {
  return (
    <Link to={to} className={`flex items-center gap-2 px-3 py-2 rounded-xl ${active ? 'bg-brand-50 text-brand-700' : 'hover:bg-gray-100'}`}>
      <Server size={16}/>
      {label}
    </Link>
  )
}

