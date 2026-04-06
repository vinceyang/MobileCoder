import { useEffect, useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getDeviceSessions, Session } from '../services/device'

export default function DeviceDetailPage() {
  const { deviceId } = useParams<{ deviceId: string }>()
  const [sessions, setSessions] = useState<Session[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const navigate = useNavigate()

  const loadSessions = useCallback(async () => {
    if (!deviceId) return
    try {
      setError('')
      const data = await getDeviceSessions(deviceId)
      setSessions(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load sessions')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [deviceId])

  useEffect(() => {
    loadSessions()
    // Auto-refresh sessions every 3 seconds to detect reconnection
    const interval = setInterval(loadSessions, 3000)
    return () => clearInterval(interval)
  }, [loadSessions])

  const connectSession = (sessionName: string) => {
    localStorage.setItem('device_id', deviceId!)
    localStorage.setItem('session_name', sessionName)
    navigate(`/terminal?device_id=${deviceId}&session_name=${encodeURIComponent(sessionName)}`)
  }

  return (
    <div className="min-h-screen bg-gray-900">
      <header className="flex items-center p-4 border-b border-gray-800">
        <button onClick={() => navigate('/devices')} className="text-gray-400 mr-4 text-xl">
          ←
        </button>
        <div>
          <h1 className="text-lg font-bold">设备详情</h1>
          <p className="text-xs text-gray-400">{deviceId}</p>
        </div>
      </header>

      <div className="p-4">
        <h2 className="text-lg font-medium mb-3">Sessions</h2>

        {loading ? (
          <p className="text-gray-400">加载中...</p>
        ) : error ? (
          <div className="text-center text-red-400 mt-4">
            <p>{error}</p>
            <button onClick={loadSessions} className="mt-2 text-blue-400 underline">
              重试
            </button>
          </div>
        ) : sessions.length === 0 ? (
          <p className="text-gray-400 text-center mt-4">暂无活跃 Session</p>
        ) : (
          <div className="space-y-3">
            {sessions.map((session) => (
              <div
                key={session.id}
                onClick={() => connectSession(session.session_name)}
                className="bg-gray-800 p-4 rounded-lg cursor-pointer active:bg-gray-700"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium">{session.session_name}</h3>
                    <p className="text-sm text-gray-400">{session.project_path}</p>
                  </div>
                  <span className={`px-2 py-1 rounded text-xs ${
                    session.status === 'active' ? 'bg-green-600' : 'bg-gray-600'
                  }`}>
                    {session.status === 'active' ? '活跃' : '非活跃'}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
