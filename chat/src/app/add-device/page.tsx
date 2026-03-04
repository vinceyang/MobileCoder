'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';

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

    console.log('Binding device with bindCode:', bindCode);

    // Create an AbortController with timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000);

    try {
      console.log('Making bind request to: /api/device/bind');
      const res = await fetch('/api/device/bind', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token,
        },
        body: JSON.stringify({ bind_code: bindCode }),
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      console.log('Response status:', res.status);
      let data;
      const text = await res.text();
      console.log('Response text:', text);
      try {
        data = JSON.parse(text);
      } catch (e) {
        // Server returned non-JSON response
        console.error('Failed to parse JSON:', e);
        setError(text || '绑定失败，请检查绑定码是否正确');
        setLoading(false);
        return;
      }
      console.log('Response data:', data);

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
    <div className="min-h-screen flex items-center justify-center bg-gray-900">
      <div className="bg-gray-800 p-8 rounded-lg w-96">
        <h1 className="text-2xl text-white mb-6">添加设备</h1>

        {error && (
          <div className="bg-red-500 text-white p-3 rounded mb-4">
            {error}
          </div>
        )}

        <p className="text-gray-400 mb-4">
          请在 Agent 端启动后获取绑定码，然后在此输入绑定码绑定设备。
        </p>

        <input
          type="text"
          placeholder="输入绑定码"
          value={bindCode}
          onChange={(e) => setBindCode(e.target.value)}
          className="w-full p-3 mb-6 bg-gray-700 text-white rounded"
        />

        <button
          onClick={handleBind}
          disabled={loading}
          className="w-full bg-blue-600 text-white p-3 rounded hover:bg-blue-700 disabled:opacity-50"
        >
          {loading ? '绑定中...' : '绑定'}
        </button>

        <button
          onClick={() => router.push('/devices')}
          className="w-full mt-4 text-gray-400 hover:text-white"
        >
          返回设备列表
        </button>
      </div>
    </div>
  );
}
