import { useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import DevicesPage from './pages/DevicesPage'
import DeviceDetailPage from './pages/DeviceDetailPage'
import TerminalPage from './pages/TerminalPage'
import TasksPage from './pages/TasksPage'
import TaskDetailPage from './pages/TaskDetailPage'
import NotificationsPage from './pages/NotificationsPage'
import { NotificationRuntime } from './components/NotificationBell'
import { onAuthExpired } from './services/api'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token')
  if (!token) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

function AuthExpirationListener() {
  const navigate = useNavigate()
  useEffect(() => onAuthExpired(() => navigate('/login', { replace: true })), [navigate])
  return null
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthExpirationListener />
      <NotificationRuntime />
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/devices" element={
          <ProtectedRoute>
            <DevicesPage />
          </ProtectedRoute>
        } />
        <Route path="/tasks" element={
          <ProtectedRoute>
            <TasksPage />
          </ProtectedRoute>
        } />
        <Route path="/tasks/:taskId" element={
          <ProtectedRoute>
            <TaskDetailPage />
          </ProtectedRoute>
        } />
        <Route path="/notifications" element={
          <ProtectedRoute>
            <NotificationsPage />
          </ProtectedRoute>
        } />
        <Route path="/devices/:deviceId" element={
          <ProtectedRoute>
            <DeviceDetailPage />
          </ProtectedRoute>
        } />
        <Route path="/terminal" element={
          <ProtectedRoute>
            <TerminalPage />
          </ProtectedRoute>
        } />
        <Route path="/" element={<Navigate to="/tasks" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
