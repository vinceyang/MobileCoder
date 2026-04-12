'use client';

import { useEffect, useMemo, useState } from 'react';
import { useRouter } from 'next/navigation';
import { getTasks, Task, TaskState } from '@/lib/tasks';
import { TaskCard } from '@/components/tasks/task-card';
import { NotificationBell } from '@/components/notifications/notification-bell';

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

  async function loadTasks() {
    try {
      setError('');
      const data = await getTasks();
      setTasks(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch tasks');
    } finally {
      setLoading(false);
    }
  }

  const visibleTasks = useMemo(() => {
    if (activeFilter === 'all') return tasks;
    return tasks.filter((task) => task.state === activeFilter);
  }, [activeFilter, tasks]);

  const handleLogout = () => {
    localStorage.clear();
    router.push('/login');
  };

  const openTask = (task: Task) => {
    router.push(`/tasks/${encodeURIComponent(task.id)}`);
  };

  return (
    <div className="min-h-screen bg-gray-900 p-4 md:p-6">
      <div className="max-w-4xl mx-auto">
        <header className="flex items-start justify-between gap-4 mb-8">
          <div>
            <p className="text-sm text-sky-300 uppercase tracking-[0.2em]">Control Tower</p>
            <h1 className="text-3xl font-semibold text-white mt-2">Tasks</h1>
            <p className="text-gray-400 mt-2">先看正在发生的工作，再决定要不要进入终端。</p>
          </div>
          <div className="flex items-center gap-3">
            <NotificationBell />
            <button
              onClick={() => router.push('/devices')}
              className="px-4 py-2 rounded-xl bg-gray-800 text-gray-200 hover:bg-gray-700"
            >
              设备
            </button>
            <button
              onClick={handleLogout}
              className="px-4 py-2 rounded-xl text-gray-400 hover:text-white"
            >
              退出
            </button>
          </div>
        </header>

        <div className="flex flex-wrap gap-2 mb-6">
          {filters.map((filter) => (
            <button
              key={filter.key}
              onClick={() => setActiveFilter(filter.key)}
              className={`px-3 py-2 rounded-full text-sm border ${
                activeFilter === filter.key
                  ? 'bg-sky-500 text-slate-950 border-sky-400'
                  : 'bg-gray-800 text-gray-300 border-gray-700 hover:bg-gray-700'
              }`}
            >
              {filter.label}
            </button>
          ))}
        </div>

        {loading ? (
          <div className="text-gray-400">加载任务中...</div>
        ) : error ? (
          <div className="bg-rose-500/10 border border-rose-500/20 text-rose-200 rounded-2xl p-4">
            <p>{error}</p>
            <button onClick={() => void loadTasks()} className="mt-3 text-sm underline">
              重试
            </button>
          </div>
        ) : visibleTasks.length === 0 ? (
          <div className="bg-gray-800 border border-gray-700 rounded-2xl p-8 text-center">
            <p className="text-white text-lg">当前没有匹配的任务</p>
            <p className="text-gray-400 mt-2">打开设备页查看 Session，或等待 Agent 创建新任务。</p>
          </div>
        ) : (
          <div className="space-y-4">
            {visibleTasks.map((task) => (
              <TaskCard key={task.id} task={task} onOpen={openTask} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
