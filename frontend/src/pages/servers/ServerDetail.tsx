import { useParams } from 'react-router-dom'

export default function ServerDetail() {
  const { id } = useParams()
  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Server: {id}</h1>
      <div className="bg-white rounded-2xl shadow-card p-6">
        <div className="text-gray-600">Config summary and deployment status will appear here.</div>
      </div>
    </div>
  )
}

