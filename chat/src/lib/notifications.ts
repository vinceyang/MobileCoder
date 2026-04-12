import { getApiBaseUrl } from './api'

export type NotificationEventType =
  | 'task_completed'
  | 'task_waiting_for_input'
  | 'task_idle_too_long'
  | 'agent_disconnected'

export interface NotificationRecord {
  id: number
  user_id: number
  task_id: string
  device_id: string
  session_name: string
  event_type: NotificationEventType | string
  title: string
  body: string
  dedupe_key: string
  read_at: string | null
  created_at: string
}

export interface NotificationQuery {
  since?: string | Date
  unread?: boolean
  limit?: number
}

export interface NotificationResponse {
  notifications: NotificationRecord[]
}

export const notificationEventLabels: Record<NotificationEventType, string> = {
  task_completed: '任务已完成',
  task_waiting_for_input: '需要确认',
  task_idle_too_long: '可能卡住',
  agent_disconnected: 'Agent 断开',
}

export const notificationEventStyles: Record<NotificationEventType, string> = {
  task_completed: 'border-cyan-400/30 bg-cyan-500/10 text-cyan-200',
  task_waiting_for_input: 'border-amber-400/30 bg-amber-500/10 text-amber-200',
  task_idle_too_long: 'border-sky-400/30 bg-sky-500/10 text-sky-200',
  agent_disconnected: 'border-rose-400/30 bg-rose-500/10 text-rose-200',
}

function isNotificationEventType(value: string): value is NotificationEventType {
  return value in notificationEventLabels
}

function getToken() {
  return localStorage.getItem('token') || ''
}

function normalizeSinceValue(value?: string | Date): string | undefined {
  if (!value) {
    return undefined
  }

  if (value instanceof Date) {
    return value.toISOString()
  }

  return value
}

export function buildNotificationRequestPath(query: NotificationQuery = {}): string {
  const params = new URLSearchParams()
  const since = normalizeSinceValue(query.since)
  if (since) {
    params.set('since', since)
  }
  if (typeof query.unread === 'boolean') {
    params.set('unread', String(query.unread))
  }
  if (typeof query.limit === 'number' && query.limit > 0) {
    params.set('limit', String(query.limit))
  }

  const suffix = params.toString()
  return suffix ? `/api/notifications?${suffix}` : '/api/notifications'
}

export async function getNotifications(query: NotificationQuery = {}): Promise<NotificationRecord[]> {
  const token = getToken()
  const res = await fetch(`${getApiBaseUrl()}${buildNotificationRequestPath(query)}`, {
    headers: { Authorization: token },
  })
  if (!res.ok) {
    throw new Error('Failed to fetch notifications')
  }

  const data = (await res.json()) as NotificationResponse
  return data.notifications || []
}

export async function markNotificationRead(notificationId: number): Promise<void> {
  const token = getToken()
  const res = await fetch(`${getApiBaseUrl()}/api/notifications/read`, {
    method: 'POST',
    headers: {
      Authorization: token,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ notification_id: notificationId }),
  })

  if (!res.ok) {
    throw new Error('Failed to mark notification read')
  }
}

export async function markAllNotificationsRead(): Promise<void> {
  const token = getToken()
  const res = await fetch(`${getApiBaseUrl()}/api/notifications/read-all`, {
    method: 'POST',
    headers: { Authorization: token },
  })

  if (!res.ok) {
    throw new Error('Failed to mark notifications read')
  }
}

export async function getUnreadNotificationCount(): Promise<number> {
  const notifications = await getNotifications({ unread: true, limit: 100 })
  return notifications.length
}

export function buildNotificationTaskHref(notification: NotificationRecord): string {
  if (!notification.task_id) {
    return '/notifications'
  }

  return `/tasks/${encodeURIComponent(notification.task_id)}`
}

export function isUnreadNotification(notification: NotificationRecord): boolean {
  return !notification.read_at
}

export function getNotificationEventLabel(eventType: string): string {
  return isNotificationEventType(eventType) ? notificationEventLabels[eventType] : eventType
}

export function getNotificationEventStyle(eventType: string): string {
  return isNotificationEventType(eventType)
    ? notificationEventStyles[eventType]
    : 'border border-slate-700 bg-slate-900/80 text-slate-400'
}

export function countUnreadNotifications(notifications: NotificationRecord[]): number {
  return notifications.filter(isUnreadNotification).length
}

function parseNotificationTime(value: string): number {
  const timestamp = Date.parse(value)
  return Number.isNaN(timestamp) ? Number.NEGATIVE_INFINITY : timestamp
}

export function sortNotificationsByFreshness(notifications: NotificationRecord[]): NotificationRecord[] {
  return [...notifications].sort((left, right) => {
    const unreadDelta = Number(isUnreadNotification(right)) - Number(isUnreadNotification(left))
    if (unreadDelta !== 0) {
      return unreadDelta
    }

    const timeDelta = parseNotificationTime(right.created_at) - parseNotificationTime(left.created_at)
    if (timeDelta !== 0) {
      return timeDelta
    }

    return right.id - left.id
  })
}

export function formatNotificationTime(value: string): string {
  if (!value) {
    return '暂无时间'
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
