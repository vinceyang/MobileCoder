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

  return (
    <button
      type="button"
      onClick={() => onOpen(task)}
      className="w-full text-left bg-gray-800 rounded-2xl p-5 hover:bg-gray-750 transition-colors border border-gray-700"
    >
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <h3 className="text-white text-lg font-semibold truncate">{task.title}</h3>
            <span className="text-[11px] uppercase tracking-wide text-sky-300 bg-sky-500/10 px-2 py-1 rounded-full border border-sky-500/20">
              {task.tool}
            </span>
          </div>
          <p className="text-gray-400 text-sm mt-1 truncate">
            {task.project_path || task.device_name}
          </p>
        </div>
        <span className={`text-xs px-2.5 py-1 rounded-full whitespace-nowrap ${stateStyles[task.state]}`}>
          {stateLabels[task.state]}
        </span>
      </div>

      <div className="mt-4 flex items-center gap-2 flex-wrap">
        <span className={`text-[11px] px-2 py-1 rounded-full ${eventKindStyles[latestEventKind]}`}>
          {eventKindLabels[latestEventKind]}
        </span>
        <p className="text-gray-200 text-sm min-w-0 flex-1">{task.recent_event || task.summary}</p>
      </div>

      <div className="flex items-center justify-between mt-4 text-xs text-gray-400">
        <span>{task.device_name}</span>
        <span>{task.session_name}</span>
      </div>
    </button>
  );
}
