import { useEffect, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { getDevices, Device } from '../services/device'

export default function DevicesPage() {
  const [devices, setDevices] = useState<Device[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const navigate = useNavigate()

  const loadDevices = useCallback(async () => {
    try {
      setError('')
      const data = await getDevices()
      setDevices(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载设备失败')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    loadDevices()
    return () => controller.abort()
  }, [loadDevices])

  const handleLogout = () => {
    localStorage.clear()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-gray-950">
      <header className="flex items-start justify-between gap-4 px-4 py-3 border-b border-gray-900/80 bg-gray-950/80 backdrop-blur-sm">
        <div className="min-w-0">
          <p className="text-[10px] font-medium text-gray-500 uppercase tracking-[0.28em]">设备管理</p>
          <h1 className="mt-1 text-lg font-semibold tracking-[0.01em] text-gray-100">设备</h1>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => navigate('/tasks')} className="rounded-full px-3 py-1.5 text-sm text-sky-300 hover:bg-sky-300/10">
            任务总览
          </button>
          <button onClick={handleLogout} className="rounded-full px-3 py-1.5 text-sm text-gray-400 hover:bg-gray-800">
            退出
          </button>
        </div>
      </header>

      <div className="p-4">
        {loading ? (
          <p className="text-center text-sm text-gray-500">加载中...</p>
        ) : error ? (
          <div className="mt-8 text-center text-red-400">
            <p>{error}</p>
            <button onClick={loadDevices} className="mt-2 text-sm text-blue-400 underline underline-offset-4">
              重试
            </button>
          </div>
        ) : devices.length === 0 ? (
          <div className="mt-8 text-center text-gray-500">
            <p>暂无设备</p>
            <p className="mt-2 text-sm text-gray-600">请在桌面 Agent 中绑定设备</p>
          </div>
        ) : (
          <div className="space-y-3">
            {devices.map((device) => (
              <div
                key={device.device_id}
                onClick={() => navigate(`/devices/${device.device_id}`)}
                className="cursor-pointer rounded-xl border border-gray-800/80 bg-transparent px-4 py-3 transition-colors hover:border-gray-700 active:bg-gray-900/50"
              >
                <div className="flex items-center justify-between">
                  <div className="min-w-0">
                    <h3 className="truncate text-sm font-medium text-gray-100">{device.device_name}</h3>
                    <p className="mt-1 truncate text-xs text-gray-500">{device.device_id}</p>
                  </div>
                  <span
                    className={`shrink-0 rounded-full border px-2 py-1 text-[11px] ${
                      device.status === 'online'
                        ? 'border-green-500/30 bg-green-500/10 text-green-300'
                        : 'border-gray-700 bg-gray-900/60 text-gray-400'
                    }`}
                  >
                    {device.status === 'online' ? '在线' : '离线'}
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
