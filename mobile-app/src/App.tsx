import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import DevicesPage from './pages/DevicesPage'
import DeviceDetailPage from './pages/DeviceDetailPage'
import TerminalPage from './pages/TerminalPage'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token')
  if (!token) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/devices" element={
          <ProtectedRoute>
            <DevicesPage />
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
        <Route path="/" element={<Navigate to="/devices" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
