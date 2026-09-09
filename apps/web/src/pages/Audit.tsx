import { useQuery } from '@tanstack/react-query'
import { ScrollText } from 'lucide-react'
import API from '@/lib/api'
import { Sidebar } from './Users'

export default function Audit() {
  const { data: entries, isLoading } = useQuery({
    queryKey: ['audit'],
    queryFn: () => API.get('/audit').then((res) => res.data),
  })

  const actionColors: Record<string, string> = {
    'login': 'text-green-400', 'login_failed': 'text-red-400',
    'create': 'text-blue-400', 'delete': 'text-red-400', 'backup': 'text-yellow-400',
  }

  return (
    <div className="min-h-screen bg-bg-base text-gray-100">
      <header className="h-14 border-b border-border bg-bg-card flex items-center px-6"><span className="font-bold text-accent">SambaForge</span></header>
      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6">
          <h1 className="text-xl font-semibold flex items-center gap-2 mb-6"><ScrollText className="h-5 w-5 text-accent" />Registro de Auditoría</h1>
          <div className="bg-bg-card border border-border rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-bg-elevated text-gray-400">
                <tr><th className="text-left px-4 py-3 font-medium">Timestamp</th><th className="text-left px-4 py-3 font-medium">Usuario</th><th className="text-left px-4 py-3 font-medium">Acción</th><th className="text-left px-4 py-3 font-medium">Recurso</th><th className="text-left px-4 py-3 font-medium">Detalle</th><th className="text-left px-4 py-3 font-medium">IP</th></tr>
              </thead>
              <tbody>
                {isLoading ? <tr><td colSpan={6} className="text-center py-8 text-gray-500">Cargando...</td></tr> :
                  entries?.map((e: any, i: number) => (
                    <tr key={i} className="border-t border-border hover:bg-bg-elevated">
                      <td className="px-4 py-3 font-mono text-xs text-gray-500">{e.timestamp}</td>
                      <td className="px-4 py-3 font-mono">{e.user}</td>
                      <td className={`px-4 py-3 font-mono ${actionColors[e.action] || 'text-gray-400'}`}>{e.action}</td>
                      <td className="px-4 py-3">{e.resource}</td>
                      <td className="px-4 py-3 text-gray-500 max-w-xs truncate">{e.detail}</td>
                      <td className="px-4 py-3 font-mono text-xs text-gray-500">{e.ip}</td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        </main>
      </div>
    </div>
  )
}