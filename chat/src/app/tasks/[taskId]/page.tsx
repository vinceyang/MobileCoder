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
    <div className="min-h-screen bg-gray-900 p-4 md:p-6">
      <div className="max-w-4xl mx-auto">
        <div className="mb-6 flex items-center justify-between gap-3">
          <button onClick={() => router.push('/tasks')} className="text-gray-400 hover:text-white">
            ← 返回任务列表
          </button>
          <NotificationBell className="shrink-0" />
        </div>

        <section className="bg-gray-800 border border-gray-700 rounded-3xl p-6 md:p-8">
          <div className="flex items-start justify-between gap-4 flex-wrap">
            <div className="min-w-0">
              <p className="text-sm text-sky-300 uppercase tracking-[0.2em]">Task Detail</p>
              <h1 className="text-3xl font-semibold text-white mt-2">{task.title}</h1>
              <p className="text-gray-400 mt-3 max-w-2xl">{task.summary}</p>
            </div>
            <span className={`text-sm px-3 py-1.5 rounded-full whitespace-nowrap ${stateStyles[task.state]}`}>
              {stateLabels[task.state]}
            </span>
          </div>

          <div className="grid md:grid-cols-2 gap-4 mt-8">
            <InfoCard label="状态原因" value={task.state_reason} />
            <InfoCard label="最近事件" value={task.recent_event} />
            <InfoCard label="项目路径" value={task.project_path || '未提供'} />
            <InfoCard label="工具" value={task.tool} />
            <InfoCard label="设备" value={task.device_name} />
            <InfoCard label="Session" value={task.session_name} />
          </div>

          <div className="mt-6 rounded-2xl border border-gray-700 bg-gray-900/60 p-5">
            <p className="text-xs uppercase tracking-wide text-gray-500">最近活动</p>
            <p className="text-base text-gray-100 mt-2">
              {formatActivityLabel(task.last_activity_at)}
            </p>
            <p className="text-sm text-gray-400 mt-2">
              这一步先使用设备在线时间和 session 创建时间做弱信号聚合，后续再接真正的 agent 事件流。
            </p>
          </div>

          <div className="mt-6 rounded-2xl border border-gray-700 bg-gray-900/60 p-5">
            <div className="flex items-center justify-between gap-3">
              <div>
                <p className="text-xs uppercase tracking-wide text-gray-500">Recent Timeline</p>
                <p className="text-sm text-gray-400 mt-2">来自实时 terminal output 的最近事件摘录。</p>
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
                      <p className="text-xs text-gray-500 mt-1">{formatActivityLabel(event.timestamp)}</p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="mt-5 rounded-2xl border border-dashed border-gray-700 p-4 text-sm text-gray-400">
                还没有捕获到实时事件，当前仍回退使用状态摘要。
              </div>
            )}
          </div>

          <div className="mt-8 flex flex-wrap gap-3">
            <button
              onClick={openTerminal}
              className="px-5 py-3 rounded-2xl bg-sky-400 text-slate-950 font-medium hover:bg-sky-300"
            >
              打开终端接管
            </button>
            <button
              onClick={() => void loadTask(task.id)}
              className="px-5 py-3 rounded-2xl bg-gray-700 text-gray-200 hover:bg-gray-600"
            >
              刷新状态
            </button>
          </div>
        </section>
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

function InfoCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-gray-700 bg-gray-900/60 p-4">
      <p className="text-xs uppercase tracking-wide text-gray-500">{label}</p>
      <p className="text-sm text-gray-100 mt-2 break-all">{value}</p>
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
