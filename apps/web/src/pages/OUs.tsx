import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { FolderTree, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import API from '@/lib/api'
import { Sidebar } from './Users'

export default function OUs() {
  const queryClient = useQueryClient()
  const [showCreate, setShowCreate] = useState(false)
  const [newOU, setNewOU] = useState('')

  const { data: ous, isLoading } = useQuery({
    queryKey: ['ous'],
    queryFn: () => API.get('/ous').then((res) => res.data),
  })

  const createMutation = useMutation({
    mutationFn: (name: string) => API.post('/ous', { name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ous'] })
      setShowCreate(false)
      setNewOU('')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (name: string) => API.delete(`/ous/${encodeURIComponent(name)}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['ous'] }),
  })

  return (
    <div className="min-h-screen bg-bg-base text-gray-100">
      <header className="h-14 border-b border-border bg-bg-card flex items-center px-6">
        <span className="font-bold text-accent">SambaForge</span>
      </header>
      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6">
          <div className="flex items-center justify-between mb-6">
            <h1 className="text-xl font-semibold">Unidades Organizativas</h1>
            <button
              onClick={() => setShowCreate(true)}
              className="flex items-center gap-2 bg-accent hover:bg-accent-hover text-white px-4 py-2 rounded-lg text-sm transition-colors"
            >
              <Plus className="h-4 w-4" />
              Nueva OU
            </button>
          </div>

          <div className="bg-bg-card border border-border rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-bg-elevated text-gray-400">
                <tr>
                  <th className="text-left px-4 py-3 font-medium">Unidad Organizativa</th>
                  <th className="text-right px-4 py-3 font-medium">Acciones</th>
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <tr><td colSpan={2} className="text-center py-8 text-gray-500">Cargando...</td></tr>
                ) : ous?.map((ou: any) => (
                  <tr key={ou.name} className="border-t border-border hover:bg-bg-elevated">
                    <td className="px-4 py-3 flex items-center gap-2">
                      <FolderTree className="h-4 w-4 text-yellow-500" />
                      <span className="font-mono">{ou.name}</span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => deleteMutation.mutate(encodeURIComponent(ou.name))}
                        className="p-1.5 text-gray-500 hover:text-red-400"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {showCreate && (
            <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowCreate(false)}>
              <div className="bg-bg-card border border-border rounded-xl p-6 w-96" onClick={(e) => e.stopPropagation()}>
                <h2 className="text-lg font-semibold mb-4">Nueva OU</h2>
                <input
                  type="text"
                  placeholder="OU=Nombre"
                  value={newOU}
                  onChange={(e) => setNewOU(e.target.value)}
                  className="w-full bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent mb-4"
                />
                <div className="flex gap-2 justify-end">
                  <button onClick={() => setShowCreate(false)} className="px-4 py-2 text-sm text-gray-400">Cancelar</button>
                  <button
                    onClick={() => createMutation.mutate(newOU)}
                    disabled={!newOU || createMutation.isPending}
                    className="px-4 py-2 bg-accent text-white rounded-lg text-sm disabled:opacity-50"
                  >
                    {createMutation.isPending ? 'Creando...' : 'Crear'}
                  </button>
                </div>
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  )
}