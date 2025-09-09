import { useEffect, useState } from 'react'

type Toast = { id: number, message: string }

export function useToast() {
  const [toasts, setToasts] = useState<Toast[]>([])
  const push = (message: string) => setToasts(t => [...t, { id: Date.now(), message }])
  const remove = (id: number) => setToasts(t => t.filter(x => x.id !== id))
  return { toasts, push, remove }
}

export function Toaster() {
  const [toasts, setToasts] = useState<Toast[]>([])
  useEffect(() => {
    ;(window as any).__pushToast = (msg: string) => setToasts(t => [...t, { id: Date.now(), message: msg }])
  }, [])
  return (
    <div className="fixed right-4 bottom-4 space-y-2 z-50">
      {toasts.map(t => (
        <div key={t.id} className="bg-gray-900 text-white px-4 py-2 rounded-xl shadow-card">
          {t.message}
        </div>
      ))}
    </div>
  )
}

export default Toaster

