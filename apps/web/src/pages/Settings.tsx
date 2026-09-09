import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ShieldCheck, QrCode, X, Loader2 } from 'lucide-react'
import { useState } from 'react'
import API from '@/lib/api'
import { useAuth } from '@/stores/auth'
import { Sidebar } from './Users'

export default function Settings() {
  const queryClient = useQueryClient()
  const { username } = useAuth()
  const [qrUrl, setQrUrl] = useState('')
  const [code, setCode] = useState('')

  const { data: status, isLoading } = useQuery({
    queryKey: ['2fa-status', username],
    queryFn: () => API.get(`/2fa/status/${username}`).then((res) => res.data),
  })

  const setupMutation = useMutation({
    mutationFn: () => API.post('/2fa/setup', { username, realm: '' }).then((res) => res.data),
    onSuccess: (data) => setQrUrl(data.qrUrl),
  })

  const verifyMutation = useMutation({
    mutationFn: (code: string) => API.post('/2fa/verify', { username, code }).then((res) => res.data),
    onSuccess: (data) => {
      if (data.valid) {
        alert('2FA verificado correctamente')
        setQrUrl('')
        queryClient.invalidateQueries({ queryKey: ['2fa-status', username] })
      }
    },
  })

  const disableMutation = useMutation({
    mutationFn: () => API.delete(`/2fa/${username}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['2fa-status', username] }),
  })

  return (
    <div className="min-h-screen bg-bg-base text-gray-100">
      <header className="h-14 border-b border-border bg-bg-card flex items-center px-6"><span className="font-bold text-accent">SambaForge</span></header>
      <div className="flex">
        <Sidebar />
        <main className="flex-1 p-6">
          <h1 className="text-xl font-semibold mb-6 flex items-center gap-2"><ShieldCheck className="h-5 w-5 text-accent" />Configuración</h1>

          {/* 2FA Section */}
          <div className="bg-bg-card border border-border rounded-xl p-6 mb-6 max-w-2xl">
            <h2 className="text-sm font-medium text-gray-400 mb-4">Autenticación de doble factor (2FA)</h2>

            {isLoading ? <p className="text-gray-500">Cargando...</p> : status?.enabled ? (
              <div className="space-y-4">
                <div className="flex items-center gap-2 text-green-400"><ShieldCheck className="h-5 w-5" />2FA está habilitado para {username}</div>
                <button onClick={() => disableMutation.mutate()} disabled={disableMutation.isPending} className="flex items-center gap-2 text-sm text-red-400 hover:text-red-300">
                  <X className="h-4 w-4" />Deshabilitar 2FA
                </button>
              </div>
            ) : qrUrl ? (
              <div className="space-y-4">
                <p className="text-sm text-gray-400">Escanea este código QR con Google Authenticator o similar:</p>
                <div className="bg-white p-4 rounded-lg inline-block">
                  <img src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(qrUrl)}`} alt="QR Code" className="w-48 h-48" />
                </div>
                <p className="text-xs text-gray-500 font-mono break-all">{qrUrl}</p>
                <div className="flex gap-2">
                  <input type="text" placeholder="Código de 6 dígitos" value={code} onChange={(e) => setCode(e.target.value)} maxLength={6} className="bg-bg-input border border-border rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:border-accent w-40" />
                  <button onClick={() => verifyMutation.mutate(code)} disabled={code.length !== 6 || verifyMutation.isPending} className="bg-accent text-white px-4 py-2 rounded-lg text-sm disabled:opacity-50 flex items-center gap-2">
                    {verifyMutation.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : null}Verificar
                  </button>
                </div>
                {verifyMutation.isError && <p className="text-red-400 text-sm">Código inválido</p>}
              </div>
            ) : (
              <div className="space-y-4">
                <p className="text-sm text-gray-400">2FA no está habilitado. Habilita doble factor para mayor seguridad.</p>
                <button onClick={() => setupMutation.mutate()} disabled={setupMutation.isPending} className="flex items-center gap-2 bg-accent hover:bg-accent-hover text-white px-4 py-2 rounded-lg text-sm">
                  {setupMutation.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <QrCode className="h-4 w-4" />}Habilitar 2FA
                </button>
              </div>
            )}
          </div>

          {/* Change Password Section */}
          <ChangePasswordSection username={username} />
        </main>
      </div>
    </div>
  )
}

function ChangePasswordSection(_props: { username: string | null }) {
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirm, setConfirm] = useState('')

  const changeMutation = useMutation({
    mutationFn: (data: { oldPassword: string; newPassword: string }) => API.post('/auth/change-password', data).then((res) => res.data),
    onSuccess: () => { setOldPassword(''); setNewPassword(''); setConfirm(''); },
  })

  return (
    <div className="bg-bg-card border border-border rounded-xl p-6 max-w-2xl">
      <h2 className="text-sm font-medium text-gray-400 mb-4">Cambiar mi contraseña</h2>
      <div className="space-y-4">
        <div>
          <label className="text-sm text-gray-400">Contraseña actual</label>
          <input type="password" value={oldPassword} onChange={(e) => setOldPassword(e.target.value)} className="w-full bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent mt-1" />
        </div>
        <div>
          <label className="text-sm text-gray-400">Nueva contraseña</label>
          <input type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} className="w-full bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent mt-1" />
        </div>
        <div>
          <label className="text-sm text-gray-400">Confirmar nueva contraseña</label>
          <input type="password" value={confirm} onChange={(e) => setConfirm(e.target.value)} className="w-full bg-bg-input border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-accent mt-1" />
        </div>
        {changeMutation.isSuccess && <p className="text-green-400 text-sm">Contraseña cambiada correctamente</p>}
        {changeMutation.isError && <p className="text-red-400 text-sm">{(changeMutation.error as any)?.response?.data?.error || 'Error al cambiar contraseña'}</p>}
        <button
          onClick={() => changeMutation.mutate({ oldPassword, newPassword })}
          disabled={!oldPassword || !newPassword || newPassword !== confirm || changeMutation.isPending}
          className="bg-accent hover:bg-accent-hover text-white px-4 py-2 rounded-lg text-sm disabled:opacity-50 flex items-center gap-2"
        >
          {changeMutation.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : null}Cambiar contraseña
        </button>
      </div>
    </div>
  )
}