import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { getDevices, Device } from '../services/device'

export default function DevicesPage() {
  const [devices, setDevices] = useState<Device[]>([])
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    loadDevices()
  }, [])

  const loadDevices = async () => {
    try {
      const data = await getDevices()
      setDevices(data)
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = () => {
    localStorage.clear()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-gray-900">
      <header className="flex items-center justify-between p-4 border-b border-gray-800">
        <h1 className="text-xl font-bold">我的设备</h1>
        <button onClick={handleLogout} className="text-gray-400 text-sm">
          退出
        </button>
      </header>

      <div className="p-4">
        {loading ? (
          <p className="text-gray-400 text-center">加载中...</p>
        ) : devices.length === 0 ? (
          <div className="text-center text-gray-400 mt-8">
            <p>暂无设备</p>
            <p className="text-sm mt-2">请在 Desktop Agent 中绑定设备</p>
          </div>
        ) : (
          <div className="space-y-3">
            {devices.map((device) => (
              <div
                key={device.device_id}
                onClick={() => navigate(`/devices/${device.device_id}`)}
                className="bg-gray-800 p-4 rounded-lg cursor-pointer active:bg-gray-700"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium">{device.device_name}</h3>
                    <p className="text-sm text-gray-400">{device.device_id}</p>
                  </div>
                  <span className={`px-2 py-1 rounded text-xs ${
                    device.status === 'online' ? 'bg-green-600' : 'bg-gray-600'
                  }`}>
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
