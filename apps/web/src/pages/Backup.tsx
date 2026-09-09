import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { HardDriveDownload, Trash2, Loader2 } from 'lucide-react'
import API from '@/lib/api'
import { Sidebar } from './Users'

export default function Backup() {
  const queryClient = useQueryClient()

  const { data: backups } = useQuery({
    queryKey: ['backups'],
    queryFn: () => API.get('/backups').then((res) => res.data),
    refetchInterval: 10000,
  })

  const backupMutation = useMutation({
    mutationFn: () => API.post('/backup', {}, { timeout: 120000 }).then((res) => res.data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['backups'] }),
  })

  const deleteMutation = useMutation({
    mutationFn: (path: string) => API.delete('/backups', { data: { path } }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['backups'] }),
  })

  return (
    <div className="min-h-screen bg-bg-base text-gray-100">
      <header className="h-14 border-b border-border bg-bg-card flex items-center px-6"><span className="font-bold text-accent">SambaForge</span></header>
      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6">
          <div className="flex items-center justify-between mb-6">
            <h1 className="text-xl font-semibold flex items-center gap-2"><HardDriveDownload className="h-5 w-5 text-accent" />Backup y Restore</h1>
            <button
              onClick={() => backupMutation.mutate()}
              disabled={backupMutation.isPending}
              className="flex items-center gap-2 bg-accent hover:bg-accent-hover text-white px-4 py-2 rounded-lg text-sm transition-colors disabled:opacity-50"
            >
              {backupMutation.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <HardDriveDownload className="h-4 w-4" />}
              {backupMutation.isPending ? 'Creando backup...' : 'Crear Backup'}
            </button>
          </div>

          {backupMutation.isError && <div className="bg-red-500/10 border border-red-500/30 text-red-400 text-sm rounded-lg p-3 mb-4">{(backupMutation.error as any)?.response?.data?.error || 'Error al crear backup'}</div>}
          {backupMutation.isSuccess && <div className="bg-green-500/10 border border-green-500/30 text-green-400 text-sm rounded-lg p-3 mb-4">Backup creado: {backupMutation.data?.file} ({backupMutation.data?.size} bytes)</div>}

          <div className="bg-bg-card border border-border rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-bg-elevated text-gray-400">
                <tr><th className="text-left px-4 py-3 font-medium">Archivo</th><th className="text-left px-4 py-3 font-medium">Tamaño</th><th className="text-left px-4 py-3 font-medium">Fecha</th><th className="text-right px-4 py-3 font-medium">Acciones</th></tr>
              </thead>
              <tbody>
                {!backups || backups.length === 0 ? <tr><td colSpan={4} className="text-center py-8 text-gray-500">No hay backups</td></tr> :
                  backups.map((b: any, i: number) => (
                    <tr key={i} className="border-t border-border hover:bg-bg-elevated">
                      <td className="px-4 py-3 font-mono text-xs">{b.file}</td>
                      <td className="px-4 py-3">{b.sizeHuman}</td>
                      <td className="px-4 py-3 text-gray-500">{b.date}</td>
                      <td className="px-4 py-3 text-right"><button onClick={() => deleteMutation.mutate(b.path)} className="p-1.5 text-gray-500 hover:text-red-400"><Trash2 className="h-4 w-4" /></button></td>
                    </tr>
                  ))
                }
              </tbody>
            </table>
          </div>
        </main>
      </div>
    </div>
  )
}