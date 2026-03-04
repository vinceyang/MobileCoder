'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';

interface Device {
  ID: number;
  DeviceID: string;
  DeviceName: string;
  Status: string;
}

export default function DevicesPage() {
  const [devices, setDevices] = useState<Device[]>([]);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }

    fetchDevices();
  }, []);

  const fetchDevices = async () => {
    const token = localStorage.getItem('token') || '';

    console.log('Fetching devices with token:', token);

    try {
      const res = await fetch('/api/devices', {
        headers: { 'Authorization': token },
      });
      console.log('Devices response status:', res.status);
      const data = await res.json();
      console.log('Devices data:', data);
      setDevices(data.devices || []);
    } catch (err) {
      console.error('Fetch devices error:', err);
    }
    setLoading(false);
  };

  const handleLogout = () => {
    localStorage.clear();
    router.push('/login');
  };

  if (loading) return <div className="text-white">加载中...</div>;

  return (
    <div className="min-h-screen bg-gray-900 p-4">
      <div className="max-w-2xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl text-white">我的设备</h1>
          <div className="space-x-4">
            <button onClick={() => router.push('/add-device')} className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
              添加设备
            </button>
            <button onClick={handleLogout} className="text-gray-400 hover:text-white">
              退出登录
            </button>
          </div>
        </div>

        {devices.length === 0 ? (
          <div className="text-center mt-8">
            <p className="text-gray-400 mb-4">暂无设备</p>
            <button onClick={() => router.push('/add-device')} className="bg-blue-600 text-white px-6 py-3 rounded hover:bg-blue-700">
              立即添加设备
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            {devices.map((device) => {
              console.log('Rendering device:', device);
              return (
                <div
                  key={device.ID}
                  className="bg-gray-800 p-4 rounded-lg cursor-pointer hover:bg-gray-700"
                  onClick={() => router.push(`/devices/${device.DeviceID}`)}
                >
                  <div className="flex justify-between items-center">
                    <div>
                      <h3 className="text-white text-lg">{device.DeviceName || '未命名'}</h3>
                      <p className="text-gray-400 text-sm">{device.DeviceID}</p>
                    </div>
                    <span className={`px-3 py-1 rounded ${device.Status === 'online' ? 'bg-green-600' : 'bg-gray-600'}`}>
                      {device.Status === 'online' ? '在线' : '离线'}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
