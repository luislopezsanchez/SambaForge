import { useQuery } from '@tanstack/react-query'
import { FolderTree, Search } from 'lucide-react'
import { useState } from 'react'
import API from '@/lib/api'
import { Sidebar } from './Users'

export default function Groups() {
  const [search, setSearch] = useState('')

  const { data: groups, isLoading } = useQuery({
    queryKey: ['groups'],
    queryFn: () => API.get('/groups').then((res) => res.data),
  })

  const filtered = groups?.filter((g: any) => g.name.toLowerCase().includes(search.toLowerCase())) || []

  return (
    <div className="min-h-screen bg-bg-base text-gray-100">
      <header className="h-14 border-b border-border bg-bg-card flex items-center px-6">
        <span className="font-bold text-accent">SambaForge</span>
      </header>
      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6">
          <h1 className="text-xl font-semibold mb-6">Grupos</h1>

          <div className="relative mb-4">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Buscar grupos..."
              className="w-full bg-bg-input border border-border rounded-lg pl-10 pr-3 py-2 text-sm focus:outline-none focus:border-accent"
            />
          </div>

          <div className="bg-bg-card border border-border rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-bg-elevated text-gray-400">
                <tr>
                  <th className="text-left px-4 py-3 font-medium">Grupo</th>
                  <th className="text-left px-4 py-3 font-medium">Tipo</th>
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <tr><td colSpan={2} className="text-center py-8 text-gray-500">Cargando...</td></tr>
                ) : (
                  filtered.map((group: any) => (
                    <tr key={group.name} className="border-t border-border hover:bg-bg-elevated">
                      <td className="px-4 py-3 flex items-center gap-2">
                        <FolderTree className="h-4 w-4 text-purple-500" />
                        <span className="font-mono">{group.name}</span>
                      </td>
                      <td className="px-4 py-3 text-gray-500">Security</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </main>
      </div>
    </div>
  )
}