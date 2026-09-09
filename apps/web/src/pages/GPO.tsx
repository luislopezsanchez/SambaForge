import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { FileText, Plus, Trash2, Shield, FolderTree, Monitor } from 'lucide-react'
import { useState } from 'react'
import API from '@/lib/api'
import { Sidebar } from './Users'

export default function GPO() {
  const queryClient = useQueryClient()
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')

  const { data: gpos, isLoading } = useQuery({
    queryKey: ['gpos'],
    queryFn: () => API.get('/gpos').then((res) => res.data),
  })
  const { data: templates } = useQuery({
    queryKey: ['gpo-templates'],
    queryFn: () => API.get('/gpos/templates').then((res) => res.data),
  })

  const createMutation = useMutation({
    mutationFn: (name: string) => API.post('/gpos', { name }),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['gpos'] }); setShowCreate(false); setNewName('') },
  })
  const deleteMutation = useMutation({
    mutationFn: (id: string) => API.delete(`/gpos/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['gpos'] }),
  })

  const templateIcons: Record<string, any> = { 'password-policy': Shield, 'drive-maps': FolderTree, 'wallpaper': Monitor, 'logon-script': FileText, 'registry-settings': FileText, 'firewall-rules': Shield }

  return (
    <div className="min-h-screen bg-bg-base text-gray-100">
      <header className="h-14 border-b border-border bg-bg-card flex items-center px-6"><span className="font-bold text-accent">SambaForge</span></header>
      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6">
          <div className="flex items-center justify-between mb-6">
            <h1 className="text-xl font-semibold flex items-center gap-2"><FileText className="h-5 w-5 text-accent" />Group Policy Objects</h1>
            <button onClick={() => setShowCreate(true)} className="flex items-center gap-2 bg-accent hover:bg-accent-hover text-white px-4 py-2 rounded-lg text-sm transition-colors">
              <Plus className="h-4 w-4" />Nuevo GPO
            </button>
          </div>

          <div className="bg-bg-card border border-border rounded-xl overflow-hidden mb-6">
            <table className="w-full text-sm">
              <thead className="bg-bg-elevated text-gray-400"><tr><th className="text-left px-4 py-3 font-medium">Nombre</th><th className="text-left px-4 py-3 font-medium">ID</th><th className="text-right px-4 py-3 font-medium">Acciones</th></tr></thead>
              <tbody>
                {isLoading ? <tr><td colSpan={3} className="text-center py-8 text-gray-500">Cargando...</td></tr> :
                  gpos?.map((g: any) => (
                    <tr key={g.id} className="border-t border-border hover:bg-bg-elevated">
                      <td className="px-4 py-3">{g.name}</td>
                      <td className="px-4 py-3 font-mono text-gray-500 text-xs">{g.id}</td>
                      <td className="px-4 py-3 text-right"><button onClick={() => deleteMutation.mutate(g.id)} className="p-1.5 text-gray-500 hover:text-red-400"><Trash2 className="h-4 w-4" /></button></td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>

          <h2 className="text-sm font-medium text-gray-400 mb-3">Plantillas preconfiguradas</h2>
          <div className="grid grid-cols-3 gap-4">
            {templates?.map((t: any) => {
              const Icon = templateIcons[t.id] || FileText
              return (
                <div key={t.id} className="bg-bg-card border border-border rounded-xl p-4 hover:border-accent cursor-pointer transition-colors">
                  <Icon className="h-6 w-6 text-accent mb-3" />
                  <h3 className="font-medium mb-1">{t.name}</h3>
                  <p className="text-xs text-gray-500">{t.description}</p>
                  <span className="inline-block mt-2 text-xs px-2 py-0.5 rounded bg-bg-elevated text-gray-400">{t.category}</span>
                </div>
              )
            })}
          </div>

          {showCreate && (
            <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowCreate(false)}>
              <div className="bg-bg-card border border-border rounded-xl p-6 w-96" onClick={(e) => e.stopPropagation()}>
                <h2 className="text-lg font-semibold mb-4">Nuevo GPO</h2>
                <input type="text" placeholder="Nombre del GPO" value={newName} onChange={(e) => setNewName(e.target.value)} className="w-full bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent mb-4" />
                <div className="flex gap-2 justify-end">
                  <button onClick={() => setShowCreate(false)} className="px-4 py-2 text-sm text-gray-400">Cancelar</button>
                  <button onClick={() => createMutation.mutate(newName)} disabled={!newName || createMutation.isPending} className="px-4 py-2 bg-accent text-white rounded-lg text-sm disabled:opacity-50">{createMutation.isPending ? 'Creando...' : 'Crear'}</button>
                </div>
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  )
}