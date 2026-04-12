'use client';

import { useEffect, useMemo, useState } from 'react';
import { useRouter } from 'next/navigation';
import { getTasks, Task, TaskState } from '@/lib/tasks';
import { TaskCard } from '@/components/tasks/task-card';
import { NotificationBell } from '@/components/notifications/notification-bell';
import { PullToRefresh } from '@/components/pull-to-refresh';
import { LogoutConfirmButton } from '@/components/logout-confirm-button';

const filters: Array<{ key: 'all' | TaskState; label: string }> = [
  { key: 'all', label: '全部' },
  { key: 'running', label: '运行中' },
  { key: 'waiting', label: '等待输入' },
  { key: 'attention', label: '需关注' },
  { key: 'completed', label: '已完成' },
];

export default function TasksPage() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeFilter, setActiveFilter] = useState<'all' | TaskState>('all');
  const [error, setError] = useState('');
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }

    void loadTasks();
  }, [router]);

  async function loadTasks(options?: { silent?: boolean }) {
    try {
      if (!options?.silent) {
        setLoading(true);
      }
      setError('');
      const data = await getTasks();
      setTasks(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch tasks');
    } finally {
      if (!options?.silent) {
        setLoading(false);
      }
    }
  }

  const visibleTasks = useMemo(() => {
    if (activeFilter === 'all') return tasks;
    return tasks.filter((task) => task.state === activeFilter);
  }, [activeFilter, tasks]);

  const openTask = (task: Task) => {
    router.push(`/tasks/${encodeURIComponent(task.id)}`);
  };

  return (
    <PullToRefresh onRefresh={() => loadTasks({ silent: true })}>
    <div className="min-h-screen bg-[#020816] text-slate-100">
      <div className="mx-auto flex min-h-screen w-full max-w-md flex-col md:max-w-4xl">
        <header className="sticky top-0 z-20 border-b border-cyan-400/10 bg-[#020816]/95 px-4 py-3 backdrop-blur md:static md:border-0 md:bg-transparent md:px-6">
          <div className="flex items-center justify-between gap-3">
            <div className="min-w-0">
              <h1 className="text-2xl font-black tracking-tight text-slate-50">Tasks</h1>
            </div>
            <div className="flex shrink-0 items-center gap-2">
            <NotificationBell />
            <button
              onClick={() => router.push('/devices')}
                className="rounded-2xl border border-cyan-400/10 bg-slate-900/80 px-3 py-2 text-sm font-semibold text-slate-200"
            >
              设备
            </button>
            <LogoutConfirmButton className="rounded-2xl px-2 py-2 text-sm text-slate-500 hover:text-slate-200">
              退出
            </LogoutConfirmButton>
          </div>
          </div>

          <div className="mt-3 flex gap-2 overflow-x-auto pb-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          {filters.map((filter) => (
            <button
              key={filter.key}
              onClick={() => setActiveFilter(filter.key)}
                className={`shrink-0 rounded-full border px-3 py-2 text-sm font-semibold transition ${
                activeFilter === filter.key
                    ? 'border-cyan-300 bg-cyan-300 text-slate-950 shadow-[0_0_24px_rgba(103,232,249,0.28)]'
                    : 'border-cyan-400/10 bg-slate-900/80 text-slate-300'
              }`}
            >
              {filter.label}
            </button>
          ))}
          </div>
        </header>

        <main className="flex-1 px-4 py-4 md:px-6">
        {loading ? (
            <div className="rounded-[24px] border border-cyan-400/10 bg-slate-950/80 px-4 py-8 text-center text-slate-400">
              加载任务中...
            </div>
        ) : error ? (
            <div className="rounded-[24px] border border-rose-500/20 bg-rose-500/10 p-4 text-rose-200">
            <p>{error}</p>
              <button onClick={() => void loadTasks()} className="mt-3 text-sm font-semibold text-cyan-200 underline">
              重试
            </button>
          </div>
        ) : visibleTasks.length === 0 ? (
            <div className="rounded-[24px] border border-cyan-400/10 bg-slate-950/80 p-8 text-center">
              <p className="text-lg font-semibold text-slate-50">当前没有匹配的任务</p>
              <p className="mt-2 text-sm text-slate-400">打开设备页查看 Session，或等待 Agent 创建新任务。</p>
          </div>
        ) : (
            <div className="space-y-3">
            {visibleTasks.map((task) => (
              <TaskCard key={task.id} task={task} onOpen={openTask} />
            ))}
          </div>
        )}
        </main>
      </div>
    </div>
    </PullToRefresh>
  );
}
