import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Shield, Save } from 'lucide-react'
import { useState, useEffect } from 'react'
import API from '@/lib/api'
import { Sidebar } from './Users'

export default function PasswordPolicy() {
  const queryClient = useQueryClient()
  const [policy, setPolicy] = useState<any>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['password-policy'],
    queryFn: () => API.get('/password-policy').then((res) => res.data),
  })

  useEffect(() => {
    if (data) setPolicy(data)
  }, [data])

  const updateMutation = useMutation({
    mutationFn: (data: any) => API.put('/password-policy', data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['password-policy'] }),
  })

  if (isLoading || !policy) {
    return (
      <div className="min-h-screen bg-bg-base text-gray-100">
        <header className="h-14 border-b border-border bg-bg-card flex items-center px-6"><span className="font-bold text-accent">SambaForge</span></header>
        <div className="flex"><Sidebar /><main className="flex-1 p-6"><p className="text-gray-500">Cargando...</p></main></div>
      </div>
    )
  }

  const fields = [
    { key: 'complexity', label: 'Complejidad', type: 'select', options: ['on', 'off'] },
    { key: 'minLength', label: 'Longitud mínima', type: 'number' },
    { key: 'historyLength', label: 'Historial de passwords', type: 'number' },
    { key: 'minAge', label: 'Edad mínima (días)', type: 'number' },
    { key: 'maxAge', label: 'Edad máxima (días)', type: 'number' },
    { key: 'lockoutDuration', label: 'Duración de bloqueo (min)', type: 'number' },
    { key: 'lockoutThreshold', label: 'Intentos antes de bloqueo', type: 'number' },
  ]

  return (
    <div className="min-h-screen bg-bg-base text-gray-100">
      <header className="h-14 border-b border-border bg-bg-card flex items-center px-6"><span className="font-bold text-accent">SambaForge</span></header>
      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6">
          <div className="flex items-center justify-between mb-6">
            <h1 className="text-xl font-semibold flex items-center gap-2"><Shield className="h-5 w-5 text-accent" />Política de Contraseña</h1>
            <button
              onClick={() => updateMutation.mutate(policy)}
              disabled={updateMutation.isPending}
              className="flex items-center gap-2 bg-accent hover:bg-accent-hover text-white px-4 py-2 rounded-lg text-sm transition-colors disabled:opacity-50"
            >
              <Save className="h-4 w-4" />
              {updateMutation.isPending ? 'Guardando...' : 'Guardar'}
            </button>
          </div>

          {updateMutation.isSuccess && <div className="bg-green-500/10 border border-green-500/30 text-green-400 text-sm rounded-lg p-3 mb-4">Política actualizada correctamente</div>}

          <div className="bg-bg-card border border-border rounded-xl p-6">
            <div className="grid grid-cols-2 gap-6">
              {fields.map((field) => (
                <div key={field.key}>
                  <label className="text-sm font-medium text-gray-400 mb-2 block">{field.label}</label>
                  {field.type === 'select' ? (
                    <select
                      value={policy[field.key] || ''}
                      onChange={(e) => setPolicy({ ...policy, [field.key]: e.target.value })}
                      className="w-full bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent"
                    >
                      {field.options?.map((opt) => <option key={opt} value={opt}>{opt}</option>)}
                    </select>
                  ) : (
                    <input
                      type="number"
                      value={policy[field.key] || ''}
                      onChange={(e) => setPolicy({ ...policy, [field.key]: parseInt(e.target.value) || 0 })}
                      className="w-full bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent"
                    />
                  )}
                </div>
              ))}
            </div>
          </div>
        </main>
      </div>
    </div>
  )
}