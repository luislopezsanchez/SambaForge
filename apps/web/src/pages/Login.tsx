import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Hammer, Lock, User, Loader2 } from 'lucide-react'
import API from '@/lib/api'
import { useAuth } from '@/stores/auth'

export default function Login() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { setAuth } = useAuth()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const { data } = await API.post('/auth/login', { username, password })
      setAuth({
        token: data.token,
        username: data.username,
        cn: data.cn,
        isAdmin: data.isAdmin,
        realm: data.realm,
      })
      navigate('/dashboard')
    } catch (err: any) {
      setError(err.response?.data?.error || 'Error de conexión')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-bg-base">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center">
          <div className="inline-flex items-center gap-2 text-3xl font-bold text-accent">
            <Hammer className="h-8 w-8" />
            <span>SambaForge</span>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="bg-bg-card border border-border rounded-xl p-8 space-y-6">
          {error && (
            <div className="bg-red-500/10 border border-red-500/30 text-red-400 text-sm rounded-lg p-3">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-400">{t('login.username')}</label>
            <div className="relative">
              <User className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500" />
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full bg-bg-input border border-border rounded-lg pl-10 pr-3 py-2.5 text-sm focus:outline-none focus:border-accent"
                placeholder="administrator"
                required
              />
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-400">{t('login.password')}</label>
            <div className="relative">
              <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500" />
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full bg-bg-input border border-border rounded-lg pl-10 pr-3 py-2.5 text-sm focus:outline-none focus:border-accent"
                placeholder="••••••••••••"
                required
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-accent hover:bg-accent-hover text-white font-medium rounded-lg py-2.5 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
          >
            {loading && <Loader2 className="h-4 w-4 animate-spin" />}
            {loading ? '...' : t('login.submit')}
          </button>
        </form>
      </div>
    </div>
  )
}