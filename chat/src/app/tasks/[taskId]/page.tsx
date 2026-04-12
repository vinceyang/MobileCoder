'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { getTask, Task, TaskEventKind } from '@/lib/tasks';
import { NotificationBell } from '@/components/notifications/notification-bell';

const stateStyles: Record<Task['state'], string> = {
  running: 'bg-emerald-600/20 text-emerald-300 border border-emerald-500/30',
  waiting: 'bg-amber-500/20 text-amber-200 border border-amber-400/30',
  completed: 'bg-slate-700 text-slate-200 border border-slate-600',
  attention: 'bg-rose-600/20 text-rose-200 border border-rose-400/30',
};

const stateLabels: Record<Task['state'], string> = {
  running: '运行中',
  waiting: '等待输入',
  completed: '已完成',
  attention: '需关注',
};

export default function TaskDetailPage() {
  const [task, setTask] = useState<Task | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const params = useParams();
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }

    const rawTaskId = params.taskId;
    const taskId = Array.isArray(rawTaskId) ? rawTaskId[0] : rawTaskId;
    if (!taskId) {
      setError('Task not found');
      setLoading(false);
      return;
    }

    void loadTask(decodeURIComponent(taskId));
  }, [params.taskId, router]);

  async function loadTask(taskId: string) {
    try {
      setError('');
      const data = await getTask(taskId);
      setTask(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch task');
    } finally {
      setLoading(false);
    }
  }

  const openTerminal = () => {
    if (!task) {
      return;
    }

    localStorage.setItem('device_id', task.device_id);
    localStorage.setItem('session_name', task.session_name);
    router.push(
      `/terminal?device_id=${task.device_id}&session_name=${encodeURIComponent(task.session_name)}&task_id=${encodeURIComponent(task.id)}`
    );
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-900 flex items-center justify-center text-gray-300">
        加载任务中...
      </div>
    );
  }

  if (error || !task) {
    return (
      <div className="min-h-screen bg-gray-900 p-4 md:p-6">
        <div className="max-w-3xl mx-auto">
          <button onClick={() => router.push('/tasks')} className="text-gray-400 hover:text-white mb-6">
            ← 返回任务列表
          </button>
          <div className="bg-rose-500/10 border border-rose-500/20 text-rose-200 rounded-2xl p-5">
            <p>{error || 'Task not found'}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#020816] text-slate-100">
      <div className="mx-auto flex min-h-screen w-full max-w-md flex-col md:max-w-4xl">
        <div className="sticky top-0 z-20 flex items-center justify-between gap-3 border-b border-cyan-400/10 bg-[#020816]/95 px-4 py-3 backdrop-blur md:static md:border-0 md:bg-transparent md:px-6">
          <button onClick={() => router.push('/tasks')} className="rounded-full border border-cyan-400/10 bg-slate-900/80 px-3 py-2 text-sm text-slate-300">
            ← 返回任务列表
          </button>
          <NotificationBell className="shrink-0" />
        </div>

        <main className="flex-1 px-4 py-4 md:px-6">
        <section className="rounded-[28px] border border-cyan-400/10 bg-[radial-gradient(circle_at_top_right,rgba(14,165,233,0.20),transparent_34%),linear-gradient(180deg,rgba(15,23,42,0.96),rgba(2,8,22,0.96))] p-4 shadow-[0_24px_80px_rgba(0,0,0,0.35)] md:p-8">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h1 className="truncate text-2xl font-black tracking-tight text-slate-50 md:text-3xl">
                {task.title}
              </h1>
            </div>
            <span className={`shrink-0 rounded-full px-3 py-1.5 text-sm font-semibold ${stateStyles[task.state]}`}>
              {stateLabels[task.state]}
            </span>
          </div>

          <div className="mt-5 grid grid-cols-2 gap-2 md:grid-cols-4">
            <InfoCard label="工具" value={task.tool} compact />
            <InfoCard label="设备" value={task.device_name} compact />
            <InfoCard label="最近活动" value={formatActivityLabel(task.last_activity_at)} compact />
            <InfoCard label="Session" value={task.session_name} compact />
          </div>

          <div className="mt-5 rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
            <p className="text-xs uppercase tracking-wide text-slate-500">当前信号</p>
            <p className="mt-2 text-sm font-medium leading-5 text-slate-100">{task.state_reason}</p>
            <p className="mt-2 line-clamp-2 text-sm leading-5 text-slate-400">{task.recent_event}</p>
          </div>

          <div className="mt-5 flex gap-2">
            <button
              onClick={openTerminal}
              className="flex-1 rounded-2xl bg-cyan-300 px-4 py-3 text-sm font-black text-slate-950 shadow-[0_0_26px_rgba(103,232,249,0.28)]"
            >
              打开终端接管
            </button>
            <button
              onClick={() => void loadTask(task.id)}
              className="rounded-2xl border border-cyan-400/10 bg-slate-900/80 px-4 py-3 text-sm font-semibold text-slate-200"
            >
              刷新
            </button>
          </div>
        </section>

          <section className="mt-4 rounded-[24px] border border-cyan-400/10 bg-slate-950/80 p-4 md:p-5">
            <div className="flex items-center justify-between gap-3">
              <div>
                <p className="text-[11px] uppercase tracking-[0.2em] text-slate-500">Recent Timeline</p>
                <p className="mt-1 text-sm text-slate-400">来自实时 terminal output 的最近事件摘录。</p>
              </div>
            </div>

            {task.timeline && task.timeline.length > 0 ? (
              <div className="mt-5 space-y-3">
                {task.timeline.map((event, index) => (
                  <div key={`${event.timestamp}-${index}`} className="flex gap-3">
                    <div className={`mt-1 h-2.5 w-2.5 rounded-full shrink-0 ${eventDotStyles[event.kind]}`} />
                    <div className="min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className={`text-[11px] px-2 py-1 rounded-full border ${eventBadgeStyles[event.kind]}`}>
                          {eventKindLabels[event.kind]}
                        </span>
                        <p className="text-sm text-gray-100 break-words">{event.summary}</p>
                      </div>
                      <p className="mt-1 text-xs text-slate-500">{formatActivityLabel(event.timestamp)}</p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="mt-5 rounded-2xl border border-dashed border-slate-800 p-4 text-sm text-slate-400">
                还没有捕获到实时事件，当前仍回退使用状态摘要。
              </div>
            )}
        </section>
        </main>
      </div>
    </div>
  );
}

const eventKindLabels: Record<TaskEventKind, string> = {
  info: '信息',
  needs_input: '待确认',
  error: '异常',
  test_result: '测试',
  completed: '完成',
  tool_step: '步骤',
}

const eventBadgeStyles: Record<TaskEventKind, string> = {
  info: 'bg-slate-700 text-slate-200 border-slate-600',
  needs_input: 'bg-amber-500/20 text-amber-200 border-amber-400/30',
  error: 'bg-rose-600/20 text-rose-200 border-rose-400/30',
  test_result: 'bg-emerald-600/20 text-emerald-300 border-emerald-500/30',
  completed: 'bg-cyan-500/20 text-cyan-200 border-cyan-400/30',
  tool_step: 'bg-sky-500/20 text-sky-200 border-sky-400/30',
}

const eventDotStyles: Record<TaskEventKind, string> = {
  info: 'bg-slate-400',
  needs_input: 'bg-amber-400',
  error: 'bg-rose-400',
  test_result: 'bg-emerald-400',
  completed: 'bg-cyan-400',
  tool_step: 'bg-sky-400',
}

function InfoCard({ label, value, compact }: { label: string; value: string; compact?: boolean }) {
  return (
    <div className="min-w-0 rounded-2xl border border-cyan-400/10 bg-slate-950/60 p-3">
      <p className="text-[10px] uppercase tracking-wide text-slate-500">{label}</p>
      <p className={`mt-2 min-w-0 text-slate-100 ${compact ? 'truncate text-xs' : 'break-words text-sm'}`}>
        {value}
      </p>
    </div>
  );
}

function formatActivityLabel(value: string): string {
  if (!value) {
    return '暂无活动时间';
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleString('zh-CN', {
    hour12: false,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}
