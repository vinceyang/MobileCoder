'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { getApiBaseUrl } from '@/lib/api';

export default function AddDevicePage() {
  const [bindCode, setBindCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }
  }, []);

  const handleBind = async () => {
    if (!bindCode) return;

    const token = localStorage.getItem('token') || '';
    if (!token) {
      setError('请先登录');
      router.push('/login');
      return;
    }

    setLoading(true);
    setError('');

    // Create an AbortController with timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000);

    try {
      const res = await fetch(`${getApiBaseUrl()}/api/device/bind`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token,
        },
        body: JSON.stringify({ bind_code: bindCode }),
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      let data;
      const text = await res.text();
      try {
        data = JSON.parse(text);
      } catch (e) {
        // Server returned non-JSON response
        console.error('Failed to parse JSON:', e);
        setError(text || '绑定失败，请检查绑定码是否正确');
        setLoading(false);
        return;
      }

      if (res.ok && data.device_id) {
        router.push('/devices');
      } else {
        setError(data.error || '绑定失败，请检查绑定码是否正确');
      }
    } catch (err) {
      console.error('Fetch error:', err);
      // More detailed error message
      let errorMsg = '网络错误，请重试';
      if (err instanceof TypeError) {
        if (err.name === 'AbortError') {
          errorMsg = '请求超时，请重试';
        } else if (err.message.includes('Failed to fetch')) {
          errorMsg = '网络错误：无法连接到服务器，请检查服务器是否运行';
        } else if (err.message.includes('NetworkError')) {
          errorMsg = '网络错误：网络请求失败';
        } else {
          errorMsg = `网络错误：${err.message}`;
        }
      }
      setError(errorMsg);
    }
    setLoading(false);
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#020816] px-4 text-slate-100">
      <div className="w-full max-w-sm rounded-[28px] border border-cyan-400/10 bg-[linear-gradient(180deg,rgba(15,23,42,0.96),rgba(2,8,22,0.96))] p-6 shadow-[0_24px_90px_rgba(0,0,0,0.42)]">
        <p className="text-[11px] uppercase tracking-[0.24em] text-cyan-300">Device Bind</p>
        <h1 className="mt-3 text-3xl font-black tracking-tight text-slate-50">添加设备</h1>
        <p className="mt-2 text-sm leading-5 text-slate-400">
          启动桌面 Agent 后，把终端里显示的 6 位绑定码填到这里。
        </p>

        {error && (
          <div className="mt-5 rounded-2xl border border-rose-400/20 bg-rose-500/10 p-3 text-sm text-rose-100">
            {error}
          </div>
        )}

        <input
          type="text"
          placeholder="输入绑定码"
          value={bindCode}
          onChange={(e) => setBindCode(e.target.value)}
          className="mt-6 w-full rounded-2xl border border-cyan-400/10 bg-slate-900 px-4 py-3 text-center font-mono text-xl uppercase tracking-[0.24em] text-white outline-none placeholder:text-sm placeholder:tracking-normal placeholder:text-slate-600 focus:border-cyan-300/50"
        />

        <button
          onClick={handleBind}
          disabled={loading}
          className="mt-5 w-full rounded-2xl bg-cyan-300 p-3 font-black text-slate-950 transition active:scale-[0.99] disabled:opacity-50"
        >
          {loading ? '绑定中...' : '绑定'}
        </button>

        <button
          onClick={() => router.push('/devices')}
          className="mt-4 w-full rounded-2xl border border-cyan-400/10 bg-slate-900/80 p-3 text-sm font-semibold text-slate-300"
        >
          返回设备列表
        </button>
      </div>
    </div>
  );
}
