'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { getApiBaseUrl } from '@/lib/api';
import { LogoutConfirmButton } from '@/components/logout-confirm-button';

interface Session {
  ID: number;
  DeviceID: string;
  SessionName: string;
  ProjectPath: string;
  Status: string;
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
    const token = localStorage.getItem('token') || '';

    try {
      const res = await fetch(`${getApiBaseUrl()}/api/devices/sessions?device_id=${deviceId}`, {
        headers: { 'Authorization': token },
      });
      const data = await res.json();
      setSessions(data.sessions || []);
    } catch (err) {
      console.error(err);
    }
    setLoading(false);
  };

  const connectSession = (sessionName: string) => {
    // 跳转到终端页面
    localStorage.setItem('device_id', deviceId);
    localStorage.setItem('session_name', sessionName);
    // 传递 API URL、device_id 和 session_name
    router.push(`/terminal?device_id=${deviceId}&session_name=${encodeURIComponent(sessionName)}`);
  };

  if (loading) return <div className="min-h-screen bg-[#020816] p-4 text-slate-400">加载中...</div>;

  return (
    <div className="min-h-screen bg-[#020816] text-slate-100">
      <div className="mx-auto flex min-h-screen w-full max-w-md flex-col md:max-w-2xl">
        <header className="sticky top-0 z-20 border-b border-cyan-400/10 bg-[#020816]/95 px-4 py-3 backdrop-blur md:static md:border-0 md:bg-transparent">
          <div className="flex items-center justify-between gap-3">
            <div className="min-w-0 flex-1">
              <button onClick={() => router.push('/devices')} className="rounded-full border border-cyan-400/10 bg-slate-900/80 px-3 py-2 text-sm text-slate-300">
                ← 返回
              </button>
              <p className="mt-2 truncate font-mono text-xs text-slate-500">{deviceId}</p>
          </div>
            <LogoutConfirmButton className="rounded-2xl px-2 py-2 text-sm text-slate-500">
            退出登录
          </LogoutConfirmButton>
        </div>
        </header>

        <main className="flex-1 px-4 py-4">
        <h2 className="mb-3 text-[11px] uppercase tracking-[0.22em] text-cyan-300">Sessions</h2>

        {sessions.length === 0 ? (
          <div className="mt-8 rounded-[24px] border border-cyan-400/10 bg-slate-950/80 p-8 text-center text-slate-400">
            暂无活跃 Session
          </div>
        ) : (
          <div className="space-y-3">
            {sessions.map((session) => (
              <div
                key={session.ID}
                className="cursor-pointer rounded-[24px] border border-cyan-400/10 bg-slate-950/80 p-4 transition active:scale-[0.99]"
                onClick={() => connectSession(session.SessionName)}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <h3 className="truncate text-lg font-black text-slate-50">{session.SessionName}</h3>
                    <p className="mt-1 truncate font-mono text-xs text-slate-500">{session.ProjectPath}</p>
                  </div>
                  <span className={`shrink-0 rounded-full px-3 py-1 text-xs font-semibold ${session.Status === 'active' ? 'bg-emerald-500/20 text-emerald-200 border border-emerald-400/30' : 'bg-slate-800 text-slate-400 border border-slate-700'}`}>
                    {session.Status === 'active' ? '活跃' : '非活跃'}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
        </main>
      </div>
    </div>
  );
}
