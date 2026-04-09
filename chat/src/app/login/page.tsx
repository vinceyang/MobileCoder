'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { getApiBaseUrl } from '@/lib/api';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isRegister, setIsRegister] = useState(false);
  const [error, setError] = useState('');
  const router = useRouter();

  // 检查是否已登录
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      router.push('/devices');
    }
  }, [router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    const endpoint = isRegister ? '/api/auth/register' : '/api/auth/login';

    try {
      const res = await fetch(`${getApiBaseUrl()}${endpoint}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });

      const data = await res.json();

      if (!res.ok) {
        setError(data.error || '操作失败');
        return;
      }

      // 保存 token 和用户信息
      localStorage.setItem('token', data.token);
      localStorage.setItem('user_id', String(data.user_id));
      localStorage.setItem('email', data.email);

      router.push('/devices');
    } catch {
      setError('网络错误');
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user_id');
    localStorage.removeItem('email');
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-900">
      <form onSubmit={handleSubmit} className="bg-gray-800 p-8 rounded-lg w-96 max-w-[90vw]">
        <h1 className="text-2xl text-white mb-6">
          {isRegister ? '注册' : '登录'}
        </h1>

        {error && (
          <div className="bg-red-500 text-white p-3 rounded mb-4">
            {error}
          </div>
        )}

        <input
          type="email"
          placeholder="邮箱"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="w-full p-3 mb-4 bg-gray-700 text-white rounded"
          required
        />

        <input
          type="password"
          placeholder="密码"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="w-full p-3 mb-6 bg-gray-700 text-white rounded"
          required
        />

        <button
          type="submit"
          className="w-full bg-blue-600 text-white p-3 rounded hover:bg-blue-700"
        >
          {isRegister ? '注册' : '登录'}
        </button>

        <p className="text-gray-400 mt-4 text-center cursor-pointer" onClick={() => setIsRegister(!isRegister)}>
          {isRegister ? '已有账号？登录' : '没有账号？注册'}
        </p>

        <button
          type="button"
          onClick={handleLogout}
          className="w-full mt-4 text-gray-500 text-sm hover:text-gray-400"
        >
          退出登录
        </button>
      </form>
    </div>
  );
}
