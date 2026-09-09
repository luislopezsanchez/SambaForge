import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Server, Users, FolderTree, Monitor, FileText, Activity } from 'lucide-react'
import axios from 'axios'

interface HealthResponse {
  status: string
  service: string
  version: string
}

export default function Dashboard() {
  const { t } = useTranslation()

  const { data: health, isLoading } = useQuery({
    queryKey: ['health'],
    queryFn: () => axios.get<HealthResponse>('/api/health').then((res) => res.data),
    refetchInterval: 10000,
  })

  const stats = [
    { label: t('dashboard.dcStatus'), value: health?.status === 'ok' ? 'OK' : '—', icon: Server, color: 'text-green-500' },
    { label: t('dashboard.users'), value: '—', icon: Users, color: 'text-blue-500' },
    { label: t('dashboard.groups'), value: '—', icon: FolderTree, color: 'text-purple-500' },
    { label: t('dashboard.computers'), value: '—', icon: Monitor, color: 'text-orange-500' },
    { label: t('dashboard.gpos'), value: '—', icon: FileText, color: 'text-cyan-500' },
  ]

  return (
    <div className="min-h-screen bg-bg-base text-gray-100">
      {/* Topbar */}
      <header className="h-14 border-b border-border bg-bg-card flex items-center px-6 justify-between">
        <div className="flex items-center gap-2 font-bold text-accent">
          <Activity className="h-5 w-5" />
          SambaForge
        </div>
        <div className="text-sm text-gray-500">
          {isLoading ? 'Connecting...' : `v${health?.version || '—'}`}
        </div>
      </header>

      {/* Content */}
      <div className="flex">
        {/* Sidebar */}
        <nav className="w-60 border-r border-border bg-bg-card min-h-[calc(100vh-3.5rem)] p-4 space-y-1">
          {['Dashboard', 'Usuarios', 'Grupos', 'OUs', 'Equipos', 'DNS', 'GPO', 'Backup', 'Audit', 'Settings'].map((item, i) => (
            <div
              key={item}
              className={`px-3 py-2 rounded-lg text-sm cursor-pointer transition-colors ${
                i === 0 ? 'bg-accent text-white' : 'text-gray-400 hover:bg-bg-elevated'
              }`}
            >
              {item}
            </div>
          ))}
        </nav>

        {/* Main */}
        <main className="flex-1 p-6">
          <h1 className="text-xl font-semibold mb-6">{t('dashboard.title')}</h1>

          {/* Stat cards */}
          <div className="grid grid-cols-5 gap-4 mb-6">
            {stats.map((stat) => {
              const Icon = stat.icon
              return (
                <div key={stat.label} className="bg-bg-card border border-border rounded-xl p-4">
                  <div className="flex items-center justify-between mb-2">
                    <Icon className={`h-5 w-5 ${stat.color}`} />
                  </div>
                  <div className="text-2xl font-bold">{stat.value}</div>
                  <div className="text-xs text-gray-500 mt-1">{stat.label}</div>
                </div>
              )
            })}
          </div>

          {/* DC Status */}
          <div className="bg-bg-card border border-border rounded-xl p-6 mb-6">
            <h2 className="text-sm font-medium text-gray-400 mb-4">{t('dashboard.dcStatus')}</h2>
            <div className="grid grid-cols-4 gap-4 text-sm">
              <div>
                <span className="text-gray-500">Samba</span>
                <p className="font-mono">—</p>
              </div>
              <div>
                <span className="text-gray-500">Uptime</span>
                <p className="font-mono">—</p>
              </div>
              <div>
                <span className="text-gray-500">DNS</span>
                <p className="font-mono text-green-500">✓</p>
              </div>
              <div>
                <span className="text-gray-500">Kerberos</span>
                <p className="font-mono text-green-500">✓</p>
              </div>
            </div>
          </div>

          {/* Activity */}
          <div className="bg-bg-card border border-border rounded-xl p-6">
            <h2 className="text-sm font-medium text-gray-400 mb-4">Actividad reciente</h2>
            <div className="space-y-2 text-sm text-gray-500">
              <p>Sin actividad reciente. El dominio no está provisionado.</p>
            </div>
          </div>
        </main>
      </div>

      {/* Status bar */}
      <footer className="h-8 border-t border-border bg-bg-card flex items-center px-6 text-xs text-gray-600 gap-4">
        <span className="flex items-center gap-1.5">
          <span className="h-2 w-2 rounded-full bg-green-500"></span>
          Active
        </span>
        <span>Samba —</span>
        <span>0 users</span>
        <span>0 groups</span>
      </footer>
    </div>
  )
}