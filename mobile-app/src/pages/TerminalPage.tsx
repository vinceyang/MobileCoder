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

  return <Terminal deviceId={deviceId} />
}
