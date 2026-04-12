'use client';

import type { Task, TaskEventKind } from '@/lib/tasks';

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

const eventKindLabels: Record<TaskEventKind, string> = {
  info: '信息',
  needs_input: '待确认',
  error: '异常',
  test_result: '测试',
  completed: '完成',
  tool_step: '步骤',
};

const eventKindStyles: Record<TaskEventKind, string> = {
  info: 'bg-slate-700 text-slate-200 border border-slate-600',
  needs_input: 'bg-amber-500/20 text-amber-200 border border-amber-400/30',
  error: 'bg-rose-600/20 text-rose-200 border border-rose-400/30',
  test_result: 'bg-emerald-600/20 text-emerald-300 border border-emerald-500/30',
  completed: 'bg-cyan-500/20 text-cyan-200 border border-cyan-400/30',
  tool_step: 'bg-sky-500/20 text-sky-200 border border-sky-400/30',
};

interface TaskCardProps {
  task: Task;
  onOpen: (task: Task) => void;
}

export function TaskCard({ task, onOpen }: TaskCardProps) {
  const latestEventKind = task.timeline?.[0]?.kind || 'info';
  const projectLabel = compactPath(task.project_path || task.device_name);

  return (
    <button
      type="button"
      onClick={() => onOpen(task)}
      className="group w-full rounded-[24px] border border-cyan-400/10 bg-[linear-gradient(180deg,rgba(15,23,42,0.96),rgba(2,8,22,0.96))] p-4 text-left shadow-[0_18px_60px_rgba(0,0,0,0.28)] transition active:scale-[0.99] md:p-5"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="h-2.5 w-2.5 shrink-0 rounded-full bg-cyan-300 shadow-[0_0_18px_rgba(103,232,249,0.65)]" />
            <h3 className="min-w-0 truncate text-lg font-black tracking-tight text-slate-50">
              {task.title}
            </h3>
          </div>
          <div className="mt-2 flex min-w-0 items-center gap-2 text-xs text-slate-400">
            <span className="shrink-0 rounded-full border border-cyan-400/10 bg-cyan-400/5 px-2 py-1 text-[10px] uppercase tracking-wide text-cyan-200">
              {task.tool}
            </span>
            <span className="min-w-0 truncate font-mono">{projectLabel}</span>
          </div>
        </div>
        <span className={`shrink-0 rounded-full px-2.5 py-1 text-xs font-semibold ${stateStyles[task.state]}`}>
          {stateLabels[task.state]}
        </span>
      </div>

      <div className="mt-4 rounded-2xl border border-slate-800 bg-slate-950/70 p-3">
        <div className="mb-2 flex items-center gap-2">
          <span className={`rounded-full px-2 py-1 text-[11px] font-semibold ${eventKindStyles[latestEventKind]}`}>
          {eventKindLabels[latestEventKind]}
        </span>
        </div>
        <p className="line-clamp-2 text-sm leading-5 text-slate-200">{task.recent_event || task.summary}</p>
      </div>

      <div className="mt-3 flex min-w-0 items-center justify-between gap-3 text-[11px] text-slate-500">
        <span className="min-w-0 truncate">{task.device_name}</span>
        <span className="min-w-0 truncate text-right font-mono">{task.session_name}</span>
      </div>
    </button>
  );
}

function compactPath(value: string): string {
  if (!value) return '未提供路径';
  const parts = value.split('/').filter(Boolean);
  if (parts.length <= 2) return value;
  return `~/${parts.slice(-2).join('/')}`;
}
