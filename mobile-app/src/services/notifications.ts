import { useEffect, useState } from 'react'
import { Capacitor, type PluginListenerHandle } from '@capacitor/core'
import {
  LocalNotifications,
  type ActionPerformed,
} from '@capacitor/local-notifications'
import { getApiBaseUrl } from '../config/api'
import { getItem, setItem } from './storage'

export type NotificationEventType =
  | 'task_completed'
  | 'task_waiting_for_input'
  | 'task_idle_too_long'
  | 'agent_disconnected'

export interface NotificationItem {
  id: number
  user_id: number
  task_id: string
  device_id: string
  session_name: string
  event_type: NotificationEventType | string
  title: string
  body: string
  dedupe_key: string
  read_at: string
  created_at: string
}

interface NotificationCursor {
  lastSyncedAt: string
  deliveredIds: number[]
}

interface NotificationCenterState {
  notifications: NotificationItem[]
  unreadCount: number
  syncing: boolean
  error: string
  lastSyncedAt: string
}

const DEFAULT_LIMIT = 50
const POLL_INTERVAL_MS = 30_000
const CURSOR_STORAGE_KEY = 'mobilecoder.notifications.cursor.v1'
const DELIVERED_LIMIT = 120

const initialState: NotificationCenterState = {
  notifications: [],
  unreadCount: 0,
  syncing: false,
  error: '',
  lastSyncedAt: '',
}

let state = initialState
let cursorCache: NotificationCursor | null = null
let runtimePromise: Promise<void> | null = null
let pollTimer: ReturnType<typeof setInterval> | null = null
let visibilityHandlerAttached = false
let nativeTapListener: PluginListenerHandle | null = null
let syncInFlight: Promise<void> | null = null

const listeners = new Set<() => void>()

const notificationEventLabels: Record<string, string> = {
  task_completed: '任务完成',
  task_waiting_for_input: '等待确认',
  task_idle_too_long: '长时间无输出',
  agent_disconnected: '设备断开',
}

const notificationEventStyles: Record<string, string> = {
  task_completed: 'bg-cyan-500/20 text-cyan-200 border border-cyan-400/30',
  task_waiting_for_input: 'bg-amber-500/20 text-amber-200 border border-amber-400/30',
  task_idle_too_long: 'bg-sky-500/20 text-sky-200 border border-sky-400/30',
  agent_disconnected: 'bg-rose-600/20 text-rose-200 border border-rose-400/30',
}

const notificationEventDotStyles: Record<string, string> = {
  task_completed: 'bg-cyan-400 shadow-[0_0_16px_rgba(34,211,238,0.35)]',
  task_waiting_for_input: 'bg-amber-400 shadow-[0_0_16px_rgba(245,158,11,0.35)]',
  task_idle_too_long: 'bg-sky-400 shadow-[0_0_16px_rgba(56,189,248,0.35)]',
  agent_disconnected: 'bg-rose-400 shadow-[0_0_16px_rgba(251,113,133,0.35)]',
}

function cloneState(): NotificationCenterState {
  return {
    ...state,
    notifications: [...state.notifications],
  }
}

function notify() {
  listeners.forEach((listener) => listener())
}

function setState(next: NotificationCenterState) {
  state = next
  notify()
}

function updateState(updater: (previous: NotificationCenterState) => NotificationCenterState) {
  setState(updater(state))
}

function getToken() {
  return localStorage.getItem('token') || ''
}

function isUnread(notification: NotificationItem) {
  return !notification.read_at
}

function compareNotifications(left: NotificationItem, right: NotificationItem) {
  const leftTime = Date.parse(left.created_at)
  const rightTime = Date.parse(right.created_at)
  if (!Number.isNaN(leftTime) && !Number.isNaN(rightTime) && leftTime !== rightTime) {
    return rightTime - leftTime
  }
  return right.id - left.id
}

function normalizeNotification(notification: NotificationItem): NotificationItem {
  return {
    ...notification,
    task_id: notification.task_id || '',
    device_id: notification.device_id || '',
    session_name: notification.session_name || '',
    read_at: notification.read_at || '',
    dedupe_key: notification.dedupe_key || '',
    title: notification.title || '通知',
    body: notification.body || '',
  }
}

function mergeNotifications(existing: NotificationItem[], incoming: NotificationItem[]) {
  const byID = new Map<number, NotificationItem>()
  for (const item of existing) {
    byID.set(item.id, item)
  }
  for (const item of incoming) {
    byID.set(item.id, normalizeNotification(item))
  }
  return Array.from(byID.values()).sort(compareNotifications)
}

async function loadCursor(): Promise<NotificationCursor> {
  if (cursorCache) {
    return cursorCache
  }

  const raw = await getItem(CURSOR_STORAGE_KEY)
  if (!raw) {
    cursorCache = { lastSyncedAt: '', deliveredIds: [] }
    return cursorCache
  }

  try {
    const parsed = JSON.parse(raw) as Partial<NotificationCursor>
    cursorCache = {
      lastSyncedAt: typeof parsed.lastSyncedAt === 'string' ? parsed.lastSyncedAt : '',
      deliveredIds: Array.isArray(parsed.deliveredIds)
        ? parsed.deliveredIds.filter((value): value is number => Number.isFinite(value))
        : [],
    }
  } catch {
    cursorCache = { lastSyncedAt: '', deliveredIds: [] }
  }

  return cursorCache
}

async function saveCursor(cursor: NotificationCursor) {
  const nextCursor: NotificationCursor = {
    lastSyncedAt: cursor.lastSyncedAt,
    deliveredIds: cursor.deliveredIds.slice(-DELIVERED_LIMIT),
  }
  cursorCache = nextCursor
  await setItem(CURSOR_STORAGE_KEY, JSON.stringify(nextCursor))
}

async function fetchNotifications(options?: {
  since?: string
  unreadOnly?: boolean
  limit?: number
}): Promise<NotificationItem[]> {
  const token = getToken()
  if (!token) {
    return []
  }

  const params = new URLSearchParams()
  if (options?.since) {
    params.set('since', options.since)
  }
  if (typeof options?.unreadOnly === 'boolean') {
    params.set('unread', String(options.unreadOnly))
  }
  if (options?.limit && options.limit > 0) {
    params.set('limit', String(options.limit))
  }

  const url = `${getApiBaseUrl()}/api/notifications${params.toString() ? `?${params.toString()}` : ''}`
  const response = await fetch(url, {
    headers: {
      Authorization: token,
    },
  })
  if (!response.ok) {
    throw new Error('拉取通知失败')
  }

  const data = (await response.json()) as { notifications?: NotificationItem[] }
  return (data.notifications || []).map(normalizeNotification)
}

function localNotificationID(notification: NotificationItem) {
  if (notification.id <= 0 || !Number.isFinite(notification.id)) {
    return 1
  }
  return Math.min(2_147_483_647, Math.max(1, Math.floor(notification.id)))
}

async function deliverLocalNotification(notification: NotificationItem) {
  if (!Capacitor.isNativePlatform()) {
    return
  }

  try {
    const permission = await LocalNotifications.checkPermissions()
    if (permission.display !== 'granted') {
      const requested = await LocalNotifications.requestPermissions()
      if (requested.display !== 'granted') {
        return
      }
    }

    await LocalNotifications.schedule({
      notifications: [
        {
          id: localNotificationID(notification),
          title: notification.title,
          body: notification.body,
          schedule: { at: new Date(Date.now() + 250) },
          extra: {
            taskId: notification.task_id,
            notificationId: notification.id,
          },
        },
      ],
    })
  } catch {
    // Fall back to in-app only.
  }
}

async function attachNativeTapListener() {
  if (nativeTapListener || !Capacitor.isNativePlatform()) {
    return
  }

  try {
    nativeTapListener = await LocalNotifications.addListener(
      'localNotificationActionPerformed',
      (action: ActionPerformed) => {
        const extra = action.notification.extra as { taskId?: string } | undefined
        if (extra?.taskId) {
          window.dispatchEvent(
            new CustomEvent('mobilecoder:notification-open', {
              detail: { taskId: extra.taskId },
            }),
          )
        }
      },
    )
  } catch {
    nativeTapListener = null
  }
}

function attachVisibilityRefresh() {
  if (visibilityHandlerAttached || typeof window === 'undefined') {
    return
  }

  const trigger = () => {
    void syncNotificationCenter()
  }

  window.addEventListener('focus', trigger)
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') {
      trigger()
    }
  })
  visibilityHandlerAttached = true
}

export async function markNotificationRead(notificationID: number): Promise<void> {
  const token = getToken()
  if (!token) {
    return
  }

  const response = await fetch(`${getApiBaseUrl()}/api/notifications/read`, {
    method: 'POST',
    headers: {
      Authorization: token,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ notification_id: notificationID }),
  })
  if (!response.ok) {
    throw new Error('标记通知已读失败')
  }
}

export async function markAllNotificationsRead(): Promise<void> {
  const token = getToken()
  if (!token) {
    return
  }

  const response = await fetch(`${getApiBaseUrl()}/api/notifications/read-all`, {
    method: 'POST',
    headers: {
      Authorization: token,
    },
  })
  if (!response.ok) {
    throw new Error('标记全部已读失败')
  }
}

async function syncNotificationCenter(): Promise<void> {
  if (syncInFlight) {
    return syncInFlight
  }

  syncInFlight = (async () => {
    const token = getToken()
    if (!token) {
      state = initialState
      cursorCache = null
      notify()
      return
    }

    updateState((current) => ({ ...current, syncing: true, error: '' }))

    try {
      const cursor = await loadCursor()
      const bootstrap = cursor.lastSyncedAt === ''
      const incoming = await fetchNotifications({
        since: bootstrap ? undefined : cursor.lastSyncedAt,
        unreadOnly: false,
        limit: DEFAULT_LIMIT,
      })

      const merged = mergeNotifications(state.notifications, incoming)
      const delivered = new Set(cursor.deliveredIds)
      const freshDelivered: number[] = []
      let nextLastSyncedAt = cursor.lastSyncedAt

      for (const item of incoming) {
        if (!nextLastSyncedAt || Date.parse(item.created_at) > Date.parse(nextLastSyncedAt)) {
          nextLastSyncedAt = item.created_at
        }

        if (bootstrap) {
          freshDelivered.push(item.id)
          continue
        }
        if (!isUnread(item) || delivered.has(item.id)) {
          continue
        }

        freshDelivered.push(item.id)
        await deliverLocalNotification(item)
      }

      setState({
        notifications: merged,
        unreadCount: merged.filter(isUnread).length,
        syncing: false,
        error: '',
        lastSyncedAt: nextLastSyncedAt || cursor.lastSyncedAt,
      })

      await saveCursor({
        lastSyncedAt: nextLastSyncedAt || cursor.lastSyncedAt,
        deliveredIds: Array.from(new Set([...cursor.deliveredIds, ...freshDelivered])),
      })
    } catch (error) {
      updateState((current) => ({
        ...current,
        syncing: false,
        error: error instanceof Error ? error.message : '同步通知失败',
      }))
    } finally {
      updateState((current) => ({ ...current, syncing: false }))
      syncInFlight = null
    }
  })()

  return syncInFlight
}

export async function refreshNotifications() {
  await syncNotificationCenter()
}

export async function startNotificationRuntime() {
  if (runtimePromise) {
    return runtimePromise
  }

  runtimePromise = (async () => {
    attachVisibilityRefresh()
    await attachNativeTapListener()
    await syncNotificationCenter()

    if (pollTimer) {
      clearInterval(pollTimer)
    }
    pollTimer = setInterval(() => {
      void syncNotificationCenter()
    }, POLL_INTERVAL_MS)
  })()

  return runtimePromise
}

export function subscribeToNotificationCenter(listener: () => void) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

export function getNotificationCenterSnapshot() {
  return cloneState()
}

export function useNotificationCenter() {
  const [snapshot, setSnapshot] = useState<NotificationCenterState>(getNotificationCenterSnapshot())

  useEffect(() => {
    const unsubscribe = subscribeToNotificationCenter(() => {
      setSnapshot(getNotificationCenterSnapshot())
    })

    void startNotificationRuntime()
    return unsubscribe
  }, [])

  return {
    ...snapshot,
    refresh: refreshNotifications,
    markNotificationRead: async (notificationID: number) => {
      await markNotificationRead(notificationID)
      await syncNotificationCenter()
    },
    markAllNotificationsRead: async () => {
      await markAllNotificationsRead()
      await syncNotificationCenter()
    },
  }
}

export function notificationEventLabel(eventType: string) {
  return notificationEventLabels[eventType] || '系统提醒'
}

export function notificationEventStyle(eventType: string) {
  return notificationEventStyles[eventType] || 'bg-slate-700 text-slate-200 border border-slate-600'
}

export function notificationEventDotStyle(eventType: string) {
  return notificationEventDotStyles[eventType] || 'bg-slate-500'
}

export function formatNotificationTime(value: string) {
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
