import { useState } from 'react'

type Step = 'basic' | 'tools' | 'prompts' | 'resources' | 'review'

export default function NewServer() {
  const [step, setStep] = useState<Step>('basic')
  const [name, setName] = useState('')
  const [version, setVersion] = useState('1.0.0')
  const [description, setDescription] = useState('')
  const [oauthEnabled, setOauthEnabled] = useState(false)
  const [issuer, setIssuer] = useState('')
  const [authzEndpoint, setAuthzEndpoint] = useState('')
  const [tokenEndpoint, setTokenEndpoint] = useState('')
  const [jwksUri, setJwksUri] = useState('')
  const [scopes, setScopes] = useState('read, write')
  const [grants, setGrants] = useState('authorization_code')

  function NextBtn() {
    const order: Step[] = ['basic', 'tools', 'prompts', 'resources', 'review']
    const idx = order.indexOf(step)
    if (idx < order.length - 1) return (
      <button className="bg-brand-600 text-white px-4 py-2 rounded-xl" onClick={()=>setStep(order[idx+1])}>Next</button>
    )
    return <button className="bg-brand-600 text-white px-4 py-2 rounded-xl">Create MCP Server</button>
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Create New MCP Server</h1>
      <div className="bg-white rounded-2xl shadow-card p-6">
        <div className="flex gap-2 text-sm mb-6">
          {(['basic','tools','prompts','resources','review'] as Step[]).map(s => (
            <button key={s} onClick={()=>setStep(s)} className={`px-3 py-1.5 rounded-xl ${step===s? 'bg-brand-600 text-white' : 'bg-gray-100'}`}>{s}</button>
          ))}
        </div>

        {step==='basic' && (
          <div className="grid gap-4">
            <div>
              <label className="block text-sm font-medium mb-1">Server Name</label>
              <input value={name} onChange={e=>setName(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Version</label>
              <input value={version} onChange={e=>setVersion(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Description</label>
              <textarea value={description} onChange={e=>setDescription(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
            </div>
            <div className="pt-4">
              <label className="flex items-center gap-2"><input type="checkbox" checked={oauthEnabled} onChange={e=>setOauthEnabled(e.target.checked)} /> MCP Authorization (OAuth 2.0)</label>
            </div>
            {oauthEnabled && (
              <div className="grid md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium mb-1">Issuer URL</label>
                  <input value={issuer} onChange={e=>setIssuer(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Authorization Endpoint</label>
                  <input value={authzEndpoint} onChange={e=>setAuthzEndpoint(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Token Endpoint</label>
                  <input value={tokenEndpoint} onChange={e=>setTokenEndpoint(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">JWKS URI</label>
                  <input value={jwksUri} onChange={e=>setJwksUri(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Supported Scopes</label>
                  <input value={scopes} onChange={e=>setScopes(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
                </div>
                <div>
                  <label className="block text-sm font-medium mb-1">Grant Types</label>
                  <input value={grants} onChange={e=>setGrants(e.target.value)} className="w-full border rounded-xl px-3 py-2" />
                </div>
                <div className="md:col-span-2">
                  <button type="button" className="bg-gray-900 text-white px-4 py-2 rounded-xl">Validate Authorization Server</button>
                </div>
              </div>
            )}
          </div>
        )}

        {step!=='basic' && (
          <div className="text-gray-500">This step is scaffolded. We will wire full CRUD and validation next.</div>
        )}

        <div className="mt-6 flex justify-end">
          <NextBtn />
        </div>
      </div>
    </div>
  )
}

