import { getApiBaseUrl } from '@/lib/api'

export type TaskState = 'running' | 'waiting' | 'completed' | 'attention'

export interface TaskEvent {
  summary: string
  timestamp: string
  kind: TaskEventKind
}

export type TaskEventKind = 'info' | 'needs_input' | 'error' | 'test_result' | 'completed' | 'tool_step'

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
  const token = localStorage.getItem('token') || ''
  const res = await fetch(`${getApiBaseUrl()}/api/tasks`, {
    headers: { Authorization: token },
  })
  if (!res.ok) {
    throw new Error('Failed to fetch tasks')
  }
  const data = await res.json()
  return data.tasks || []
}

export async function getTask(taskId: string): Promise<Task> {
  const token = localStorage.getItem('token') || ''
  const res = await fetch(`${getApiBaseUrl()}/api/tasks/detail?id=${encodeURIComponent(taskId)}`, {
    headers: { Authorization: token },
  })
  if (!res.ok) {
    throw new Error('Failed to fetch task')
  }
  const data = await res.json()
  return data.task
}
