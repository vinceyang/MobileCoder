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
    <div className="h-screen flex flex-col" style={{ backgroundColor: '#111827' }}>
      {/* 合并的头部：返回 + 连接状态 + 重连 */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-gray-800" style={{ backgroundColor: '#111827' }}>
        <button
          onClick={goBack}
          style={{ color: '#9ca3af' }}
          className="hover:text-white text-lg"
        >
          ←
        </button>
        <div className="flex items-center gap-2">
          <span className={`w-2 h-2 rounded-full ${connected ? 'bg-green-500' : 'bg-red-500'}`} />
          <span className="text-xs" style={{ color: '#9ca3af' }}>
            {connected ? '已连接' : '重连中...'}
          </span>
        </div>
        <button
          onClick={reconnect}
          className="text-xs px-2 py-1 rounded"
          style={{ color: '#60a5fa', backgroundColor: 'rgba(59, 130, 246, 0.1)' }}
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
