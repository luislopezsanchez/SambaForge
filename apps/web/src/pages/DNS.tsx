import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Globe, Plus, Trash2 } from 'lucide-react'
import { useState } from 'react'
import API from '@/lib/api'
import { Sidebar } from './Users'

export default function DNS() {
  const queryClient = useQueryClient()
  const [selectedZone, setSelectedZone] = useState('')
  const [showAdd, setShowAdd] = useState(false)
  const [newRecord, setNewRecord] = useState({ zone: '', name: '', type: 'A', data: '' })

  const { data: zones } = useQuery({
    queryKey: ['dns-zones'],
    queryFn: () => API.get('/dns/zones').then((res) => res.data),
  })

  const { data: records, isLoading } = useQuery({
    queryKey: ['dns-records', selectedZone],
    queryFn: () => API.get(`/dns/zones/${selectedZone}/records?name=@`).then((res) => res.data),
    enabled: !!selectedZone,
  })

  const addMutation = useMutation({
    mutationFn: (data: typeof newRecord) => API.post('/dns/records', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dns-records', newRecord.zone] })
      setShowAdd(false)
      setNewRecord({ zone: '', name: '', type: 'A', data: '' })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (data: any) => API.delete('/dns/records', { data }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['dns-records', selectedZone] }),
  })

  return (
    <div className="min-h-screen bg-bg-base text-gray-100">
      <header className="h-14 border-b border-border bg-bg-card flex items-center px-6"><span className="font-bold text-accent">SambaForge</span></header>
      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6">
          <div className="flex items-center justify-between mb-6">
            <h1 className="text-xl font-semibold flex items-center gap-2"><Globe className="h-5 w-5 text-accent" />DNS</h1>
            <button
              onClick={() => { setNewRecord({ ...newRecord, zone: selectedZone }); setShowAdd(true) }}
              disabled={!selectedZone}
              className="flex items-center gap-2 bg-accent hover:bg-accent-hover text-white px-4 py-2 rounded-lg text-sm transition-colors disabled:opacity-50"
            >
              <Plus className="h-4 w-4" />
              Nuevo Record
            </button>
          </div>

          {/* Zone selector */}
          <div className="flex gap-2 mb-4">
            {zones?.map((zone: any) => (
              <button
                key={zone.name}
                onClick={() => setSelectedZone(zone.name)}
                className={`px-4 py-2 rounded-lg text-sm font-mono transition-colors ${selectedZone === zone.name ? 'bg-accent text-white' : 'bg-bg-card border border-border text-gray-400 hover:bg-bg-elevated'}`}
              >
                {zone.name}
              </button>
            ))}
          </div>

          {/* Records table */}
          {selectedZone && (
            <div className="bg-bg-card border border-border rounded-xl overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-bg-elevated text-gray-400">
                  <tr>
                    <th className="text-left px-4 py-3 font-medium">Nombre</th>
                    <th className="text-left px-4 py-3 font-medium">Tipo</th>
                    <th className="text-left px-4 py-3 font-medium">Datos</th>
                    <th className="text-right px-4 py-3 font-medium">Acciones</th>
                  </tr>
                </thead>
                <tbody>
                  {isLoading ? (
                    <tr><td colSpan={4} className="text-center py-8 text-gray-500">Cargando...</td></tr>
                  ) : records?.map((rec: any, i: number) => (
                    <tr key={i} className="border-t border-border hover:bg-bg-elevated">
                      <td className="px-4 py-3 font-mono">{rec.name || '@'}</td>
                      <td className="px-4 py-3"><span className="px-2 py-1 rounded bg-accent/20 text-accent text-xs font-mono">{rec.type}</span></td>
                      <td className="px-4 py-3 font-mono text-gray-300">{rec.data}</td>
                      <td className="px-4 py-3 text-right">
                        <button
                          onClick={() => deleteMutation.mutate({ zone: rec.zone, name: rec.name, type: rec.type, data: rec.data })}
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
          )}

          {!selectedZone && <p className="text-gray-500 text-sm">Selecciona una zona para ver sus records</p>}

          {/* Add Record Modal */}
          {showAdd && (
            <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowAdd(false)}>
              <div className="bg-bg-card border border-border rounded-xl p-6 w-96" onClick={(e) => e.stopPropagation()}>
                <h2 className="text-lg font-semibold mb-4">Nuevo Record DNS</h2>
                <div className="space-y-4">
                  <div>
                    <label className="text-sm text-gray-400">Zona</label>
                    <input type="text" value={newRecord.zone} onChange={(e) => setNewRecord({ ...newRecord, zone: e.target.value })} className="w-full bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent mt-1" />
                  </div>
                  <div>
                    <label className="text-sm text-gray-400">Nombre</label>
                    <input type="text" placeholder="wiki" value={newRecord.name} onChange={(e) => setNewRecord({ ...newRecord, name: e.target.value })} className="w-full bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent mt-1" />
                  </div>
                  <div>
                    <label className="text-sm text-gray-400">Tipo</label>
                    <select value={newRecord.type} onChange={(e) => setNewRecord({ ...newRecord, type: e.target.value })} className="w-full bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent mt-1">
                      {['A', 'AAAA', 'CNAME', 'MX', 'SRV', 'TXT', 'PTR', 'NS'].map(t => <option key={t} value={t}>{t}</option>)}
                    </select>
                  </div>
                  <div>
                    <label className="text-sm text-gray-400">Datos</label>
                    <input type="text" placeholder="172.30.36.50" value={newRecord.data} onChange={(e) => setNewRecord({ ...newRecord, data: e.target.value })} className="w-full bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent mt-1" />
                  </div>
                  {addMutation.isError && <div className="text-red-400 text-sm">{(addMutation.error as any)?.response?.data?.error}</div>}
                  <div className="flex gap-2 justify-end">
                    <button onClick={() => setShowAdd(false)} className="px-4 py-2 text-sm text-gray-400">Cancelar</button>
                    <button onClick={() => addMutation.mutate(newRecord)} disabled={!newRecord.name || !newRecord.data || addMutation.isPending} className="px-4 py-2 bg-accent text-white rounded-lg text-sm disabled:opacity-50">
                      {addMutation.isPending ? 'Creando...' : 'Crear'}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  )
}