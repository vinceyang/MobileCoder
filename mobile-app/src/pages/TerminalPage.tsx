import { useEffect, useState } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import Terminal from '../components/Terminal'

export default function TerminalPage() {
  const [searchParams] = useSearchParams()
  const [deviceId, setDeviceId] = useState('')
  const [taskId, setTaskId] = useState('')
  const navigate = useNavigate()

  useEffect(() => {
    const deviceIdParam = searchParams.get('device_id')
    const sessionNameParam = searchParams.get('session_name')
    const taskIdParam = searchParams.get('task_id') || ''

    const id = deviceIdParam || localStorage.getItem('device_id')
    if (!id) {
      navigate('/tasks')
      return
    }

    setDeviceId(id)
    setTaskId(taskIdParam)
    if (sessionNameParam) {
      localStorage.setItem('session_name', sessionNameParam)
    }
  }, [searchParams, navigate])

  if (!deviceId) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-950">
        <p className="text-gray-400">加载中...</p>
      </div>
    )
  }

  return (
    <div className="flex h-[100dvh] flex-col bg-gray-950">
      <header className="flex items-center gap-3 border-b border-gray-900/80 bg-gray-950/90 px-4 py-2 backdrop-blur-sm">
        <button
          onClick={() => {
            if (taskId) {
              navigate(`/tasks/${encodeURIComponent(taskId)}`)
              return
            }
            navigate(`/devices/${deviceId}`)
          }}
          className="flex items-center gap-2 rounded-full px-2 py-1 text-sm text-gray-300 hover:bg-gray-800 hover:text-white"
        >
          <span aria-hidden="true" className="text-base leading-none">←</span>
          <span>{taskId ? '返回任务' : '返回设备'}</span>
        </button>
        <div className="min-w-0">
          <p className="text-[10px] font-medium uppercase tracking-[0.28em] text-gray-500">CONTROL TOWER</p>
          <span className="mt-1 block text-xs text-gray-400">终端接管</span>
        </div>
      </header>
      <div className="flex-1 overflow-hidden">
        <Terminal deviceId={deviceId} />
      </div>
    </div>
  )
}
