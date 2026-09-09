import { create } from 'zustand'

interface AuthState {
  token: string | null
  username: string | null
  cn: string | null
  isAdmin: boolean
  realm: string | null
  setAuth: (data: { token: string; username: string; cn: string; isAdmin: boolean; realm: string }) => void
  logout: () => void
  isAuthenticated: () => boolean
}

export const useAuth = create<AuthState>((set, get) => ({
  token: localStorage.getItem('sambaforge_token'),
  username: localStorage.getItem('sambaforge_user'),
  cn: localStorage.getItem('sambaforge_cn'),
  isAdmin: localStorage.getItem('sambaforge_admin') === 'true',
  realm: localStorage.getItem('sambaforge_realm'),
  setAuth: (data) => {
    localStorage.setItem('sambaforge_token', data.token)
    localStorage.setItem('sambaforge_user', data.username)
    localStorage.setItem('sambaforge_cn', data.cn)
    localStorage.setItem('sambaforge_admin', String(data.isAdmin))
    localStorage.setItem('sambaforge_realm', data.realm)
    set(data)
  },
  logout: () => {
    localStorage.removeItem('sambaforge_token')
    localStorage.removeItem('sambaforge_user')
    localStorage.removeItem('sambaforge_cn')
    localStorage.removeItem('sambaforge_admin')
    localStorage.removeItem('sambaforge_realm')
    set({ token: null, username: null, cn: null, isAdmin: false, realm: null })
  },
  isAuthenticated: () => !!get().token,
}))