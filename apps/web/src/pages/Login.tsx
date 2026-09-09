import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Hammer, Lock, User } from 'lucide-react'

export default function Login() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    // TODO: real LDAP bind auth
    setTimeout(() => {
      setLoading(false)
      navigate('/dashboard')
    }, 500)
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-bg-base">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center">
          <div className="inline-flex items-center gap-2 text-3xl font-bold text-accent">
            <Hammer className="h-8 w-8" />
            <span>{t('login.title')}</span>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="bg-bg-card border border-border rounded-xl p-8 space-y-6">
          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-400">{t('login.username')}</label>
            <div className="relative">
              <User className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500" />
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full bg-bg-input border border-border rounded-lg pl-10 pr-3 py-2.5 text-sm focus:outline-none focus:border-accent"
                placeholder="admin@domain.com"
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
              />
            </div>
          </div>

          <label className="flex items-center gap-2 text-sm text-gray-400">
            <input type="checkbox" className="rounded border-border bg-bg-input" />
            {t('login.remember')}
          </label>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-accent hover:bg-accent-hover text-white font-medium rounded-lg py-2.5 transition-colors disabled:opacity-50"
          >
            {loading ? '...' : t('login.submit')}
          </button>
        </form>

        <p className="text-center text-xs text-gray-600">
          {t('login.noDomain')}
        </p>
      </div>
    </div>
  )
}