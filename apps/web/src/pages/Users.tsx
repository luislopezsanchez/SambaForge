import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { UserPlus, Trash2, KeyRound, Power, Search } from 'lucide-react'
import API from '@/lib/api'

export default function Users() {
  const queryClient = useQueryClient()
  const [search, setSearch] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [newUser, setNewUser] = useState({ username: '', password: '', firstName: '', lastName: '' })

  const { data: users, isLoading } = useQuery({
    queryKey: ['users'],
    queryFn: () => API.get('/users').then((res) => res.data),
  })

  const createMutation = useMutation({
    mutationFn: (data: typeof newUser) => API.post('/users', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      setShowCreate(false)
      setNewUser({ username: '', password: '', firstName: '', lastName: '' })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (username: string) => API.delete(`/users/${username}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['users'] }),
  })

  const filtered = users?.filter((u: any) => u.username.toLowerCase().includes(search.toLowerCase())) || []

  return (
    <div className="min-h-screen bg-bg-base text-gray-100">
      <header className="h-14 border-b border-border bg-bg-card flex items-center px-6">
        <span className="font-bold text-accent">SambaForge</span>
      </header>

      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6">
          <div className="flex items-center justify-between mb-6">
            <h1 className="text-xl font-semibold">Usuarios</h1>
            <button
              onClick={() => setShowCreate(true)}
              className="flex items-center gap-2 bg-accent hover:bg-accent-hover text-white px-4 py-2 rounded-lg text-sm transition-colors"
            >
              <UserPlus className="h-4 w-4" />
              Nuevo Usuario
            </button>
          </div>

          {/* Search */}
          <div className="relative mb-4">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Buscar usuarios..."
              className="w-full bg-bg-input border border-border rounded-lg pl-10 pr-3 py-2 text-sm focus:outline-none focus:border-accent"
            />
          </div>

          {/* Table */}
          <div className="bg-bg-card border border-border rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-bg-elevated text-gray-400">
                <tr>
                  <th className="text-left px-4 py-3 font-medium">Usuario</th>
                  <th className="text-left px-4 py-3 font-medium">CN</th>
                  <th className="text-right px-4 py-3 font-medium">Acciones</th>
                </tr>
              </thead>
              <tbody>
                {isLoading ? (
                  <tr><td colSpan={3} className="text-center py-8 text-gray-500">Cargando...</td></tr>
                ) : (
                  filtered.map((user: any) => (
                    <tr key={user.username} className="border-t border-border hover:bg-bg-elevated">
                      <td className="px-4 py-3 font-mono">{user.username}</td>
                      <td className="px-4 py-3 text-gray-400">{user.cn || '—'}</td>
                      <td className="px-4 py-3 text-right">
                        <div className="flex items-center justify-end gap-2">
                          <button className="p-1.5 text-gray-500 hover:text-yellow-400" title="Reset password">
                            <KeyRound className="h-4 w-4" />
                          </button>
                          <button className="p-1.5 text-gray-500 hover:text-orange-400" title="Disable/Enable">
                            <Power className="h-4 w-4" />
                          </button>
                          <button
                            onClick={() => deleteMutation.mutate(user.username)}
                            className="p-1.5 text-gray-500 hover:text-red-400"
                            title="Delete"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </main>
      </div>

      {/* Create User Modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowCreate(false)}>
          <div className="bg-bg-card border border-border rounded-xl p-6 w-96" onClick={(e) => e.stopPropagation()}>
            <h2 className="text-lg font-semibold mb-4">Nuevo Usuario</h2>
            <div className="space-y-4">
              <input
                type="text"
                placeholder="Nombre de usuario"
                value={newUser.username}
                onChange={(e) => setNewUser({ ...newUser, username: e.target.value })}
                className="w-full bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent"
              />
              <input
                type="password"
                placeholder="Contraseña"
                value={newUser.password}
                onChange={(e) => setNewUser({ ...newUser, password: e.target.value })}
                className="w-full bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent"
              />
              <div className="grid grid-cols-2 gap-3">
                <input
                  type="text"
                  placeholder="Nombre"
                  value={newUser.firstName}
                  onChange={(e) => setNewUser({ ...newUser, firstName: e.target.value })}
                  className="bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent"
                />
                <input
                  type="text"
                  placeholder="Apellido"
                  value={newUser.lastName}
                  onChange={(e) => setNewUser({ ...newUser, lastName: e.target.value })}
                  className="bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent"
                />
              </div>
              {createMutation.isError && (
                <div className="text-red-400 text-sm">{(createMutation.error as any)?.response?.data?.error}</div>
              )}
              <div className="flex gap-2 justify-end">
                <button onClick={() => setShowCreate(false)} className="px-4 py-2 text-sm text-gray-400 hover:text-gray-300">Cancelar</button>
                <button
                  onClick={() => createMutation.mutate(newUser)}
                  disabled={!newUser.username || !newUser.password || createMutation.isPending}
                  className="px-4 py-2 bg-accent hover:bg-accent-hover text-white rounded-lg text-sm disabled:opacity-50"
                >
                  {createMutation.isPending ? 'Creando...' : 'Crear'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export function Sidebar() {
  const navItems = [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Usuarios', path: '/users' },
    { label: 'Grupos', path: '/groups' },
    { label: 'OUs', path: '/ous' },
    { label: 'Equipos', path: '/computers' },
    { label: 'DNS', path: '/dns' },
    { label: 'GPOs', path: '/gpos' },
    { label: 'Política de Contraseña', path: '/password-policy' },
    { label: 'Backup', path: '/backup' },
    { label: 'Auditoría', path: '/audit' },
    { label: 'Multi-DC', path: '/multi-dc' },
    { label: 'Configuración', path: '/settings' },
  ]
  const currentPath = window.location.pathname
  return (
    <nav className="w-60 border-r border-border bg-bg-card min-h-[calc(100vh-3.5rem)] p-4 space-y-1">
      {navItems.map((item) => (
        <a
          key={item.path}
          href={item.path}
          className={`block px-3 py-2 rounded-lg text-sm transition-colors ${currentPath === item.path ? 'bg-accent text-white' : 'text-gray-400 hover:bg-bg-elevated'}`}
        >
          {item.label}
        </a>
      ))}
    </nav>
  )
}