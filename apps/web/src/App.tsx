import { Routes, Route, Navigate } from 'react-router-dom'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Users from './pages/Users'
import Groups from './pages/Groups'
import Computers from './pages/Computers'
import OUs from './pages/OUs'
import PasswordPolicy from './pages/PasswordPolicy'
import DNS from './pages/DNS'
import GPO from './pages/GPO'
import Audit from './pages/Audit'
import Backup from './pages/Backup'
import Settings from './pages/Settings'
import MultiDC from './pages/MultiDC'
import { useAuth } from './stores/auth'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth()
  if (!isAuthenticated()) {
    return <Navigate to="/" replace />
  }
  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Login />} />
      <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
      <Route path="/users" element={<ProtectedRoute><Users /></ProtectedRoute>} />
      <Route path="/groups" element={<ProtectedRoute><Groups /></ProtectedRoute>} />
      <Route path="/computers" element={<ProtectedRoute><Computers /></ProtectedRoute>} />
      <Route path="/ous" element={<ProtectedRoute><OUs /></ProtectedRoute>} />
      <Route path="/password-policy" element={<ProtectedRoute><PasswordPolicy /></ProtectedRoute>} />
      <Route path="/dns" element={<ProtectedRoute><DNS /></ProtectedRoute>} />
      <Route path="/gpos" element={<ProtectedRoute><GPO /></ProtectedRoute>} />
      <Route path="/backup" element={<ProtectedRoute><Backup /></ProtectedRoute>} />
      <Route path="/audit" element={<ProtectedRoute><Audit /></ProtectedRoute>} />
      <Route path="/multi-dc" element={<ProtectedRoute><MultiDC /></ProtectedRoute>} />
      <Route path="/settings" element={<ProtectedRoute><Settings /></ProtectedRoute>} />
    </Routes>
  )
}