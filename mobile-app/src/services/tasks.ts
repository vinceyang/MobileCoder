import { getApiBaseUrl } from '../config/api'

function getToken() {
  return localStorage.getItem('token') || ''
}

export type TaskState = 'running' | 'waiting' | 'completed' | 'attention'
export type TaskEventKind = 'info' | 'needs_input' | 'error' | 'test_result' | 'completed' | 'tool_step'

export interface TaskEvent {
  summary: string
  timestamp: string
  kind: TaskEventKind
}

export interface Task {
  id: string
  title: string
  device_id: string
  device_name: string
  session_name: string
  project_path: string
  tool: string
  state: TaskState
  summary: string
  state_reason: string
  recent_event: string
  last_activity_at: string
  timeline?: TaskEvent[]
}

export async function getTasks(): Promise<Task[]> {
  const token = getToken()
  const res = await fetch(`${getApiBaseUrl()}/api/tasks`, {
    headers: { Authorization: token },
  })
  if (!res.ok) throw new Error('Failed to fetch tasks')
  const data = await res.json()
  return data.tasks || []
}

export async function getTask(taskId: string): Promise<Task> {
  const token = getToken()
  const res = await fetch(`${getApiBaseUrl()}/api/tasks/detail?id=${encodeURIComponent(taskId)}`, {
    headers: { Authorization: token },
  })
  if (!res.ok) throw new Error('Failed to fetch task')
  const data = await res.json()
  return data.task
}

export const taskStateLabels: Record<TaskState, string> = {
  running: '运行中',
  waiting: '等待输入',
  completed: '已完成',
  attention: '需关注',
}

export const taskStateStyles: Record<TaskState, string> = {
  running: 'bg-emerald-600/20 text-emerald-300 border border-emerald-500/30',
  waiting: 'bg-amber-500/20 text-amber-200 border border-amber-400/30',
  completed: 'bg-slate-700 text-slate-200 border border-slate-600',
  attention: 'bg-rose-600/20 text-rose-200 border border-rose-400/30',
}

export const taskEventKindLabels: Record<TaskEventKind, string> = {
  info: '信息更新',
  needs_input: '待你确认',
  error: '执行异常',
  test_result: '测试结果',
  completed: '已完成',
  tool_step: '工具步骤',
}

export const taskEventKindStyles: Record<TaskEventKind, string> = {
  info: 'bg-slate-700 text-slate-200 border border-slate-600',
  needs_input: 'bg-amber-500/20 text-amber-200 border border-amber-400/30',
  error: 'bg-rose-600/20 text-rose-200 border border-rose-400/30',
  test_result: 'bg-emerald-600/20 text-emerald-300 border border-emerald-500/30',
  completed: 'bg-cyan-500/20 text-cyan-200 border border-cyan-400/30',
  tool_step: 'bg-sky-500/20 text-sky-200 border border-sky-400/30',
}

export const taskEventKindDotStyles: Record<TaskEventKind, string> = {
  info: 'bg-slate-400',
  needs_input: 'bg-amber-400 shadow-[0_0_16px_rgba(245,158,11,0.35)]',
  error: 'bg-rose-400 shadow-[0_0_16px_rgba(251,113,133,0.35)]',
  test_result: 'bg-emerald-400 shadow-[0_0_16px_rgba(52,211,153,0.35)]',
  completed: 'bg-cyan-400 shadow-[0_0_16px_rgba(34,211,238,0.35)]',
  tool_step: 'bg-sky-400 shadow-[0_0_16px_rgba(56,189,248,0.35)]',
}

export function formatActivityLabel(value: string): string {
  if (!value) {
    return '暂无活动时间'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return date.toLocaleString('zh-CN', {
    hour12: false,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
