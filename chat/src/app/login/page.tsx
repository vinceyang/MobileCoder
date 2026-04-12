'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { getApiBaseUrl } from '@/lib/api';
import { LogoutConfirmButton } from '@/components/logout-confirm-button';

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
      router.push('/tasks');
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

      router.push('/tasks');
    } catch {
      setError('网络错误');
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#020816] px-4 text-slate-100">
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-sm rounded-[28px] border border-cyan-400/10 bg-[radial-gradient(circle_at_top_right,rgba(14,165,233,0.22),transparent_35%),linear-gradient(180deg,rgba(15,23,42,0.96),rgba(2,8,22,0.96))] p-6 shadow-[0_24px_90px_rgba(0,0,0,0.42)]"
      >
        <p className="text-[11px] uppercase tracking-[0.24em] text-cyan-300">MobileCoder</p>
        <h1 className="mt-3 text-3xl font-black tracking-tight text-slate-50">
          {isRegister ? '创建账号' : '登录控制塔台'}
        </h1>
        <p className="mt-2 text-sm leading-5 text-slate-400">
          在手机上监督 AI coding 任务，必要时快速接管终端。
        </p>

        {error && (
          <div className="mt-5 rounded-2xl border border-rose-400/20 bg-rose-500/10 p-3 text-sm text-rose-100">
            {error}
          </div>
        )}

        <input
          type="email"
          placeholder="邮箱"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="mt-6 w-full rounded-2xl border border-cyan-400/10 bg-slate-900 px-4 py-3 text-white outline-none placeholder:text-slate-600 focus:border-cyan-300/50"
          required
        />

        <input
          type="password"
          placeholder="密码"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="mt-3 w-full rounded-2xl border border-cyan-400/10 bg-slate-900 px-4 py-3 text-white outline-none placeholder:text-slate-600 focus:border-cyan-300/50"
          required
        />

        <button
          type="submit"
          className="mt-5 w-full rounded-2xl bg-cyan-300 p-3 font-black text-slate-950 transition active:scale-[0.99]"
        >
          {isRegister ? '注册' : '登录'}
        </button>

        <p className="mt-4 cursor-pointer text-center text-sm text-slate-400" onClick={() => setIsRegister(!isRegister)}>
          {isRegister ? '已有账号？登录' : '没有账号？注册'}
        </p>

        <LogoutConfirmButton className="mt-4 w-full text-sm text-slate-600 hover:text-slate-400">
          退出登录
        </LogoutConfirmButton>
      </form>
    </div>
  );
}
