import { useQuery } from '@tanstack/react-query'
import { Monitor } from 'lucide-react'
import API from '@/lib/api'
import { Sidebar } from './Users'

export default function Computers() {
  const { data: computers, isLoading } = useQuery({
    queryKey: ['computers'],
    queryFn: () => API.get('/computers').then((res) => res.data),
  })

  return (
    <div className="min-h-screen bg-bg-base text-gray-100">
      <header className="h-14 border-b border-border bg-bg-card flex items-center px-6">
        <span className="font-bold text-accent">SambaForge</span>
      </header>
      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6">
          <h1 className="text-xl font-semibold mb-6">Equipos</h1>
          <div className="bg-bg-card border border-border rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-bg-elevated text-gray-400">
                <tr>
                  <th className="text-left px-4 py-3 font-medium">Equipo</th>
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <tr><td className="text-center py-8 text-gray-500">Cargando...</td></tr>
                ) : computers?.map((comp: any) => (
                  <tr key={comp.name} className="border-t border-border hover:bg-bg-elevated">
                    <td className="px-4 py-3 flex items-center gap-2">
                      <Monitor className="h-4 w-4 text-orange-500" />
                      <span className="font-mono">{comp.name}</span>
                    </td>
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