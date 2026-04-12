'use client';

import { useEffect, useState, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import Terminal from '../components/Terminal';

function TerminalContent() {
  const [deviceId, setDeviceId] = useState('');
  const [taskId, setTaskId] = useState('');
  const [connected, setConnected] = useState(false);
  const searchParams = useSearchParams();
  const router = useRouter();

  useEffect(() => {
    const deviceIdFromUrl = searchParams.get('device_id');
    const deviceIdFromStorage = localStorage.getItem('device_id');
    const deviceId = deviceIdFromUrl || deviceIdFromStorage;
    const taskIdFromUrl = searchParams.get('task_id') || '';

    if (!deviceId) {
      router.push('/login');
      return;
    }
    setDeviceId(deviceId);
    setTaskId(taskIdFromUrl);
  }, [searchParams, router]);

  const reconnect = () => {
    window.location.reload();
  };

  const goBack = () => {
    if (taskId) {
      router.push(`/tasks/${encodeURIComponent(taskId)}`);
      return;
    }
    router.back();
  };

  if (!deviceId) {
    return null;
  }

  return (
    <div className="flex h-screen flex-col bg-[#020816] text-slate-100">
      {/* 合并的头部：返回 + 连接状态 + 重连 */}
      <div className="flex items-center justify-between border-b border-cyan-400/10 bg-slate-950/95 px-3 py-2">
        <button
          onClick={goBack}
          className="flex h-11 w-11 items-center justify-center rounded-2xl border border-cyan-400/10 bg-slate-900/80 text-lg text-slate-300"
        >
          ←
        </button>
        <div className="flex items-center gap-2">
          <span className={`h-2 w-2 rounded-full ${connected ? 'bg-emerald-400 shadow-[0_0_14px_rgba(52,211,153,0.75)]' : 'bg-rose-400'}`} />
          <span className="text-xs font-semibold text-slate-400">
            {connected ? '已连接' : '重连中...'}
          </span>
        </div>
        <button
          onClick={reconnect}
          className="flex h-11 items-center rounded-2xl border border-cyan-400/10 bg-slate-900/80 px-3 text-xs font-semibold text-cyan-200"
        >
          重连
        </button>
      </div>
      <div className="flex-1 overflow-hidden">
        <Terminal deviceId={deviceId} onConnectionChange={setConnected} />
      </div>
    </div>
  );
}

export default function TerminalPage() {
  return (
    <Suspense fallback={<div className="h-screen flex items-center justify-center" style={{ backgroundColor: '#111827', color: '#ffffff' }}>加载中...</div>}>
      <TerminalContent />
    </Suspense>
  );
}
