import { Link } from 'react-router-dom'

export default function Dashboard() {
  const servers = [] as Array<{ id: string, name: string }>
  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold">Your MCP Servers</h1>
        <Link to="/servers/new" className="bg-brand-600 text-white px-4 py-2 rounded-xl">+ New Server</Link>
      </div>
      {servers.length === 0 ? (
        <div className="bg-white rounded-2xl shadow-card p-8 text-center text-gray-500">
          No servers yet. Click “New Server” to create one.
        </div>
      ) : (
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
          {servers.map(s => (
            <Link key={s.id} to={`/servers/${s.id}`} className="bg-white rounded-2xl shadow-card p-6 hover:shadow-lg transition">
              <div className="font-medium">{s.name}</div>
              <div className="text-sm text-gray-500">0 Tools • 0 Prompts • 0 Resources</div>
              <div className="mt-4"><span className="text-brand-700">Manage</span></div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}

