'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

// 获取 API 地址 - 支持手机端访问
const getAPIUrl = () => {
  if (typeof window === 'undefined') return 'http://localhost:8080';
  return process.env.NEXT_PUBLIC_API_URL || `${window.location.protocol}//${window.location.hostname}:8080`;
};

// 获取服务器地址用于显示
const getServerAddress = () => {
  if (typeof window === 'undefined') return 'localhost:8080';
  return process.env.NEXT_PUBLIC_API_URL?.replace(/^https?:\/\//, '') || `${window.location.hostname}:8080`;
};

export default function BindPage({ onBind }: { onBind: (deviceId: string) => void }) {
  const [bindCode, setBindCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [serverAddr, setServerAddr] = useState('localhost:8080');
  const router = useRouter();

  useEffect(() => {
    setServerAddr(getServerAddress());
  }, []);

  const handleBind = async () => {
    if (!bindCode) return;
    setLoading(true);
    setError('');
    try {
      const API_URL = getAPIUrl();
      // 简化版：无需 token，直接绑定
      const res = await fetch(`${API_URL}/api/device/bind`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ bind_code: bindCode }),
      });
      const data = await res.json();
      if (res.ok && data.device_id) {
        localStorage.setItem('device_id', data.device_id);
        onBind(data.device_id);
        router.refresh();
      } else {
        setError(data.error || '绑定失败，请检查绑定码是否正确');
      }
    } catch (err) {
      console.error(err);
      setError('网络错误，请重试');
    }
    setLoading(false);
  };

  return (
    <div className="min-h-[100dvh] bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 flex flex-col items-center justify-center p-4">
      {/* 背景装饰 */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-blue-500/10 rounded-full blur-3xl"></div>
        <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-purple-500/10 rounded-full blur-3xl"></div>
      </div>

      <div className="relative w-full max-w-[320px]">
        {/* 图标 */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 md:w-20 md:h-20 rounded-2xl bg-gradient-to-br from-green-500 to-emerald-600 mb-4 shadow-lg shadow-green-500/25">
            <svg className="w-8 h-8 md:w-10 md:h-10 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
            </svg>
          </div>
          <h1 className="text-2xl md:text-3xl font-bold text-white">绑定设备</h1>
          <p className="text-gray-400 text-sm mt-1">连接 Desktop Agent</p>
        </div>

        {/* 绑定卡片 */}
        <div className="bg-gray-800/50 backdrop-blur-sm border border-gray-700/50 rounded-2xl p-6 md:p-8 shadow-xl">
          <div className="text-center mb-6">
            <p className="text-gray-400 text-sm">
              1. 先启动 Desktop Agent，它会显示一个绑定码<br/>
              2. 在下方输入绑定码并绑定
            </p>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-sm text-center">
              {error}
            </div>
          )}

          <input
            type="text"
            value={bindCode}
            onChange={(e) => setBindCode(e.target.value)}
            placeholder="输入绑定码"
            className="w-full px-4 py-4 bg-gray-900/50 border border-gray-600 rounded-xl text-white text-center text-xl font-mono tracking-[0.3em] placeholder-gray-500 focus:outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 transition-all"
            maxLength={6}
            autoFocus
          />

          <button
            onClick={handleBind}
            disabled={loading || !bindCode}
            className="w-full py-3.5 mt-4 bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-600 hover:to-emerald-700 active:from-green-700 active:to-emerald-800 text-white font-medium rounded-xl transition-all transform active:scale-[0.98] disabled:opacity-50 disabled:transform-none shadow-lg shadow-green-500/25"
          >
            {loading ? (
              <span className="flex items-center justify-center gap-2">
                <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                绑定中...
              </span>
            ) : (
              '绑定设备'
            )}
          </button>
        </div>

        {/* 启动说明 */}
        <div className="mt-6 p-4 bg-gray-800/30 rounded-xl border border-gray-700/30">
          <p className="text-gray-500 text-xs text-center mb-2">Desktop Agent 启动命令:</p>
          <code className="block text-center text-gray-400 text-xs font-mono bg-gray-900/50 px-3 py-2 rounded-lg">
            ./bin/client -server {serverAddr}
          </code>
        </div>
      </div>
    </div>
  );
}
