'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { getApiBaseUrl } from '@/lib/api';
import { LogoutConfirmButton } from '@/components/logout-confirm-button';

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

    try {
      const res = await fetch(`${getApiBaseUrl()}/api/devices`, {
        headers: { 'Authorization': token },
      });
      const data = await res.json();
      setDevices(data.devices || []);
    } catch (err) {
      console.error('Fetch devices error:', err);
    }
    setLoading(false);
  };

  if (loading) return <div className="min-h-screen bg-[#020816] p-4 text-slate-400">加载中...</div>;

  return (
    <div className="min-h-screen bg-[#020816] text-slate-100">
      <div className="mx-auto flex min-h-screen w-full max-w-md flex-col md:max-w-2xl">
        <header className="sticky top-0 z-20 border-b border-cyan-400/10 bg-[#020816]/95 px-4 py-3 backdrop-blur md:static md:border-0 md:bg-transparent">
          <div className="mb-3 flex items-center justify-between gap-3">
            <h1 className="min-w-0 truncate text-2xl font-black tracking-tight text-slate-50">我的设备</h1>
          </div>
          <div className="grid grid-cols-3 gap-2">
            <button onClick={() => router.push('/tasks')} className="rounded-2xl border border-cyan-400/10 bg-slate-900/80 px-3 py-2 text-sm font-semibold text-slate-200">
              Tasks
            </button>
            <button onClick={() => router.push('/add-device')} className="rounded-2xl bg-cyan-300 px-3 py-2 text-sm font-black text-slate-950">
              添加设备
            </button>
            <LogoutConfirmButton className="rounded-2xl px-3 py-2 text-sm text-slate-500">
              退出登录
            </LogoutConfirmButton>
          </div>
        </header>

        <main className="flex-1 px-4 py-4">
        {devices.length === 0 ? (
          <div className="mt-8 rounded-[24px] border border-cyan-400/10 bg-slate-950/80 p-8 text-center">
            <p className="mb-4 text-slate-400">暂无设备</p>
            <button onClick={() => router.push('/add-device')} className="rounded-2xl bg-cyan-300 px-6 py-3 font-black text-slate-950">
              立即添加设备
            </button>
          </div>
        ) : (
          <div className="space-y-3">
            {devices.map((device) => {
              return (
                <div
                  key={device.ID}
                  className="cursor-pointer rounded-[24px] border border-cyan-400/10 bg-slate-950/80 p-4 transition active:scale-[0.99]"
                  onClick={() => router.push(`/devices/${device.DeviceID}`)}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <h3 className="truncate text-lg font-black text-slate-50">{device.DeviceName || '未命名'}</h3>
                      <p className="mt-1 truncate font-mono text-xs text-slate-500">{device.DeviceID}</p>
                    </div>
                    <span className={`shrink-0 rounded-full px-3 py-1 text-xs font-semibold ${device.Status === 'online' ? 'bg-emerald-500/20 text-emerald-200 border border-emerald-400/30' : 'bg-slate-800 text-slate-400 border border-slate-700'}`}>
                      {device.Status === 'online' ? '在线' : '离线'}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        )}
        </main>
      </div>
    </div>
  );
}
