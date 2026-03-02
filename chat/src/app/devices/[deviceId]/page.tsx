'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';

interface Session {
  id: number;
  device_id: string;
  session_name: string;
  project_path: string;
  status: string;
}

export default function DeviceDetailPage() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);
  const [deviceId, setDeviceId] = useState('');
  const router = useRouter();
  const params = useParams();

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }

    const id = params.deviceId as string;
    setDeviceId(id);
    fetchSessions(id);
  }, [params.deviceId]);

  const fetchSessions = async (deviceId: string) => {
    const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    const token = localStorage.getItem('token') || '';

    try {
      const res = await fetch(`${API_URL}/api/devices/sessions?device_id=${deviceId}`, {
        headers: { 'Authorization': token },
      });
      const data = await res.json();
      setSessions(data.sessions || []);
    } catch (err) {
      console.error(err);
    }
    setLoading(false);
  };

  const handleLogout = () => {
    localStorage.clear();
    router.push('/login');
  };

  const connectSession = (sessionId: number) => {
    // 跳转到终端页面，传入 session_id
    // 暂时使用 device_id 作为连接标识
    localStorage.setItem('device_id', deviceId);
    router.push('/');
  };

  if (loading) return <div className="text-white">加载中...</div>;

  return (
    <div className="min-h-screen bg-gray-900 p-4">
      <div className="max-w-2xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <div>
            <button onClick={() => router.push('/devices')} className="text-gray-400 hover:text-white mb-2">
              ← 返回设备列表
            </button>
            <h1 className="text-2xl text-white">设备详情</h1>
            <p className="text-gray-400 text-sm">{deviceId}</p>
          </div>
          <button onClick={handleLogout} className="text-gray-400 hover:text-white">
            退出登录
          </button>
        </div>

        <h2 className="text-xl text-white mb-4">Sessions</h2>

        {sessions.length === 0 ? (
          <div className="text-gray-400 text-center mt-8">
            暂无活跃 Session
          </div>
        ) : (
          <div className="space-y-4">
            {sessions.map((session) => (
              <div
                key={session.id}
                className="bg-gray-800 p-4 rounded-lg cursor-pointer hover:bg-gray-700"
                onClick={() => connectSession(session.id)}
              >
                <div className="flex justify-between items-center">
                  <div>
                    <h3 className="text-white text-lg">{session.session_name}</h3>
                    <p className="text-gray-400 text-sm">{session.project_path}</p>
                  </div>
                  <span className={`px-3 py-1 rounded ${session.status === 'active' ? 'bg-green-600' : 'bg-gray-600'}`}>
                    {session.status === 'active' ? '活跃' : '非活跃'}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
