import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Server, Users, FolderTree, Monitor, Activity, LogOut } from 'lucide-react'
import API from '@/lib/api'
import { useAuth } from '@/stores/auth'
import { useNavigate } from 'react-router-dom'

export default function Dashboard() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { cn, realm, logout } = useAuth()

  const { data: dashboard, isLoading } = useQuery({
    queryKey: ['dashboard'],
    queryFn: () => API.get('/dashboard').then((res) => res.data),
    refetchInterval: 30000,
  })

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  const stats = [
    { label: t('dashboard.dcStatus'), value: dashboard?.sambaActive ? '● Active' : '● Down', icon: Server, color: dashboard?.sambaActive ? 'text-green-500' : 'text-red-500' },
    { label: t('dashboard.users'), value: dashboard?.counts?.users ?? '—', icon: Users, color: 'text-blue-500' },
    { label: t('dashboard.groups'), value: dashboard?.counts?.groups ?? '—', icon: FolderTree, color: 'text-purple-500' },
    { label: t('dashboard.computers'), value: dashboard?.counts?.computers ?? '—', icon: Monitor, color: 'text-orange-500' },
  ]

  const sambaVersion = dashboard?.sambaVersion?.replace('samba-tool: missing subcommand\n\n', '').trim() || '—'

  return (
    <div className="min-h-screen bg-bg-base text-gray-100">
      {/* Topbar */}
      <header className="h-14 border-b border-border bg-bg-card flex items-center px-6 justify-between">
        <div className="flex items-center gap-2 font-bold text-accent">
          <Activity className="h-5 w-5" />
          SambaForge
        </div>
        <div className="flex items-center gap-4 text-sm">
          <span className="text-gray-400">{realm}</span>
          <span className="text-gray-500">{cn}</span>
          <button onClick={handleLogout} className="flex items-center gap-1 text-gray-400 hover:text-red-400 transition-colors">
            <LogOut className="h-4 w-4" />
            Logout
          </button>
        </div>
      </header>

      <div className="flex">
        {/* Sidebar */}
        <nav className="w-60 border-r border-border bg-bg-card min-h-[calc(100vh-3.5rem)] p-4 space-y-1">
          {[
            { label: 'Dashboard', path: '/dashboard' },
            { label: 'Usuarios', path: '/users' },
            { label: 'Grupos', path: '/groups' },
            { label: 'Equipos', path: '/computers' },
            { label: 'DNS', path: '/dns' },
          ].map((item, i) => (
            <div
              key={item.path}
              onClick={() => navigate(item.path)}
              className={`px-3 py-2 rounded-lg text-sm cursor-pointer transition-colors ${i === 0 ? 'bg-accent text-white' : 'text-gray-400 hover:bg-bg-elevated'}`}
            >
              {item.label}
            </div>
          ))}
        </nav>

        {/* Main */}
        <main className="flex-1 p-6">
          <h1 className="text-xl font-semibold mb-6">{t('dashboard.title')}</h1>

          {/* Stat cards */}
          <div className="grid grid-cols-4 gap-4 mb-6">
            {stats.map((stat) => {
              const Icon = stat.icon
              return (
                <div key={stat.label} className="bg-bg-card border border-border rounded-xl p-4">
                  <div className="flex items-center justify-between mb-2">
                    <Icon className={`h-5 w-5 ${stat.color}`} />
                  </div>
                  <div className="text-2xl font-bold">{isLoading ? '...' : stat.value}</div>
                  <div className="text-xs text-gray-500 mt-1">{stat.label}</div>
                </div>
              )
            })}
          </div>

          {/* DC Status */}
          <div className="bg-bg-card border border-border rounded-xl p-6 mb-6">
            <h2 className="text-sm font-medium text-gray-400 mb-4">Controlador de Dominio</h2>
            <div className="grid grid-cols-4 gap-4 text-sm">
              <div>
                <span className="text-gray-500">Samba</span>
                <p className="font-mono text-gray-300">{sambaVersion}</p>
              </div>
              <div>
                <span className="text-gray-500">Realm</span>
                <p className="font-mono text-gray-300">{dashboard?.realm || '—'}</p>
              </div>
              <div>
                <span className="text-gray-500">DNS</span>
                <p className={`font-mono ${dashboard?.sambaActive ? 'text-green-500' : 'text-red-500'}`}>
                  {dashboard?.sambaActive ? '✓ Active' : '✗ Down'}
                </p>
              </div>
              <div>
                <span className="text-gray-500">Forwarders</span>
                <p className="font-mono text-gray-300">{dashboard?.forwarders?.join(', ') || '—'}</p>
              </div>
            </div>
          </div>

          {/* Domain Level */}
          {dashboard?.domainLevel && (
            <div className="bg-bg-card border border-border rounded-xl p-6">
              <h2 className="text-sm font-medium text-gray-400 mb-4">Nivel de Dominio</h2>
              <pre className="font-mono text-xs text-gray-400 whitespace-pre-wrap">
                {dashboard.domainLevel}
              </pre>
            </div>
          )}
        </main>
      </div>

      {/* Status bar */}
      <footer className="h-8 border-t border-border bg-bg-card flex items-center px-6 text-xs text-gray-600 gap-4">
        <span className="flex items-center gap-1.5">
          <span className={`h-2 w-2 rounded-full ${dashboard?.sambaActive ? 'bg-green-500' : 'bg-red-500'}`}></span>
          {dashboard?.sambaActive ? 'Active' : 'Down'}
        </span>
        <span>Samba {sambaVersion}</span>
        <span>{dashboard?.counts?.users ?? 0} users</span>
        <span>{dashboard?.counts?.groups ?? 0} groups</span>
        <span>{dashboard?.counts?.computers ?? 0} computers</span>
      </footer>
    </div>
  )
}