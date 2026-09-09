import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Network, ArrowRight } from 'lucide-react'
import { useState } from 'react'
import API from '@/lib/api'
import { Sidebar } from './Users'

export default function MultiDC() {
  const queryClient = useQueryClient()
  const [transferRole, setTransferRole] = useState('schema')

  const { data: fsmo, isLoading: fsmoLoading } = useQuery({
    queryKey: ['fsmo'],
    queryFn: () => API.get('/fsmo').then((res) => res.data),
  })

  const { data: trusts } = useQuery({
    queryKey: ['trusts'],
    queryFn: () => API.get('/trusts').then((res) => res.data),
  })

  const transferMutation = useMutation({
    mutationFn: (data: { role: string; targetDC?: string }) => API.post('/fsmo/transfer', data).then((res) => res.data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['fsmo'] }),
  })

  const fsmoLabels: Record<string, string> = {
    schema: 'Schema Master', naming: 'Naming Master', pdc: 'PDC Emulator',
    rid: 'RID Manager', infrastructure: 'Infrastructure Master',
  }

  return (
    <div className="min-h-screen bg-bg-base text-gray-100">
      <header className="h-14 border-b border-border bg-bg-card flex items-center px-6"><span className="font-bold text-accent">SambaForge</span></header>
      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6">
          <h1 className="text-xl font-semibold mb-6 flex items-center gap-2"><Network className="h-5 w-5 text-accent" />Multi-DC y Dominio</h1>

          {/* FSMO Roles */}
          <div className="bg-bg-card border border-border rounded-xl p-6 mb-6 max-w-3xl">
            <h2 className="text-sm font-medium text-gray-400 mb-4">Roles FSMO</h2>
            {fsmoLoading ? <p className="text-gray-500">Cargando...</p> : (
              <div className="space-y-3">
                {Object.entries(fsmoLabels).map(([key, label]) => {
                  const value = fsmo?.[key] || '—'
                  const dcName = value.includes('CN=SAMBAFORGE') ? 'SAMBAFORGE' : value.includes('CN=') ? value.split('CN=')[2]?.split(',')[0] || value : '—'
                  return (
                    <div key={key} className="flex items-center justify-between py-2 border-b border-border last:border-0">
                      <div>
                        <span className="text-sm font-medium">{label}</span>
                        <p className="text-xs text-gray-500 font-mono">{dcName}</p>
                      </div>
                      <span className={value !== '—' ? 'text-green-400 text-xs' : 'text-gray-600 text-xs'}>
                        {value !== '—' ? '✓ Asignado' : 'No asignado'}
                      </span>
                    </div>
                  )
                })}
              </div>
            )}

            {/* Transfer FSMO */}
            <div className="mt-6 pt-4 border-t border-border">
              <h3 className="text-xs text-gray-500 mb-2">Transferir rol FSMO</h3>
              <div className="flex gap-2">
                <select value={transferRole} onChange={(e) => setTransferRole(e.target.value)} className="bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent">
                  {Object.entries(fsmoLabels).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                </select>
                <button
                  onClick={() => transferMutation.mutate({ role: transferRole })}
                  disabled={transferMutation.isPending}
                  className="flex items-center gap-2 bg-accent hover:bg-accent-hover text-white px-4 py-2 rounded-lg text-sm disabled:opacity-50"
                >
                  <ArrowRight className="h-4 w-4" />
                  {transferMutation.isPending ? 'Transfiriendo...' : 'Transferir'}
                </button>
              </div>
              {transferMutation.isSuccess && <p className="text-green-400 text-xs mt-2">Rol transferido correctamente</p>}
              {transferMutation.isError && <p className="text-red-400 text-xs mt-2">Error al transferir</p>}
            </div>
          </div>

          {/* Trusts */}
          <div className="bg-bg-card border border-border rounded-xl p-6 max-w-3xl">
            <h2 className="text-sm font-medium text-gray-400 mb-4">Trusts de dominio</h2>
            {!trusts || trusts.length === 0 ? (
              <p className="text-gray-500 text-sm">No hay trusts configurados.</p>
            ) : (
              <ul className="space-y-2">
                {trusts.map((t: string, i: number) => <li key={i} className="text-sm font-mono">{t}</li>)}
              </ul>
            )}
          </div>
        </main>
      </div>
    </div>
  )
}