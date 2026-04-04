import { useEffect, useState } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import Terminal from '../components/Terminal'

export default function TerminalPage() {
  const [searchParams] = useSearchParams()
  const [deviceId, setDeviceId] = useState('')
  const navigate = useNavigate()

  useEffect(() => {
    const deviceIdParam = searchParams.get('device_id')
    const sessionNameParam = searchParams.get('session_name')

    const id = deviceIdParam || localStorage.getItem('device_id')
    if (!id) {
      navigate('/devices')
      return
    }

    setDeviceId(id)
    if (sessionNameParam) {
      localStorage.setItem('session_name', sessionNameParam)
    }
  }, [searchParams])

  if (!deviceId) {
    return (
      <div className="min-h-screen bg-gray-900 flex items-center justify-center">
        <p className="text-gray-400">加载中...</p>
      </div>
    )
  }

  return (
    <div className="h-[100dvh] flex flex-col">
      <header className="flex items-center px-4 py-2 bg-gray-900 border-b border-gray-800">
        <button
          onClick={() => navigate(`/devices/${deviceId}`)}
          className="text-gray-400 hover:text-white text-xl mr-4"
        >
          ←
        </button>
        <span className="text-gray-400 text-sm">终端</span>
      </header>
      <div className="flex-1 overflow-hidden">
        <Terminal deviceId={deviceId} />
      </div>
    </div>
  )
}
