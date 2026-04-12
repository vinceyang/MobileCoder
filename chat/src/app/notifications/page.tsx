'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { NotificationBell } from '@/components/notifications/notification-bell'
import { NotificationList } from '@/components/notifications/notification-list'
import { PullToRefresh } from '@/components/pull-to-refresh'
import {
  buildNotificationTaskHref,
  countUnreadNotifications,
  getNotifications,
  isUnreadNotification,
  markAllNotificationsRead,
  markNotificationRead,
  sortNotificationsByFreshness,
  type NotificationRecord,
} from '@/lib/notifications'

const filters: Array<{ key: 'all' | 'unread'; label: string }> = [
  { key: 'all', label: '全部' },
  { key: 'unread', label: '未读' },
]

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<NotificationRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [activeFilter, setActiveFilter] = useState<'all' | 'unread'>('all')
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
      return
    }

    void loadNotifications()
    const interval = window.setInterval(() => {
      void loadNotifications({ silent: true })
    }, 30000)

    return () => {
      window.clearInterval(interval)
    }
  }, [router])

  async function loadNotifications(options?: { silent?: boolean }) {
    try {
      if (!options?.silent) {
        setLoading(true)
      }
      setError('')
      const data = await getNotifications({ limit: 100 })
      setNotifications(sortNotificationsByFreshness(data))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch notifications')
    } finally {
      if (!options?.silent) {
        setLoading(false)
      }
    }
  }

  const visibleNotifications = useMemo(() => {
    if (activeFilter === 'unread') {
      return notifications.filter(isUnreadNotification)
    }
    return notifications
  }, [activeFilter, notifications])

  const unreadCount = countUnreadNotifications(notifications)

  const handleMarkAllRead = async () => {
    try {
      await markAllNotificationsRead()
      setNotifications((current) =>
        current.map((notification) => (isUnreadNotification(notification) ? { ...notification, read_at: new Date().toISOString() } : notification)),
      )
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to mark notifications read')
    }
  }

  const handleOpen = async (notification: NotificationRecord) => {
    const taskHref = buildNotificationTaskHref(notification)
    if (isUnreadNotification(notification)) {
      try {
        await markNotificationRead(notification.id)
        setNotifications((current) =>
          current.map((item) =>
            item.id === notification.id ? { ...item, read_at: item.read_at || new Date().toISOString() } : item,
          ),
        )
      } catch {
        // Keep navigation responsive even if read-state sync fails.
      }
    }
    router.push(taskHref)
  }

  return (
    <PullToRefresh onRefresh={() => loadNotifications({ silent: true })}>
    <div className="min-h-screen bg-[#020816] text-slate-100">
      <div className="mx-auto flex min-h-screen w-full max-w-md flex-col">
        <header className="border-b border-cyan-400/10 bg-slate-950/50 px-4 py-3 backdrop-blur">
          <div className="flex items-center justify-between gap-3">
            <button
              onClick={() => router.push('/tasks')}
              className="rounded-full border border-cyan-400/10 bg-slate-900/80 px-3 py-2 text-sm font-semibold text-slate-300"
            >
              ← 返回
            </button>
            <h1 className="min-w-0 flex-1 truncate text-xl font-black tracking-tight text-slate-50">通知中心</h1>
            <NotificationBell className="shrink-0" />
          </div>

          <div className="mt-3 grid grid-cols-3 gap-2">
            <SummaryPill label="未读" value={String(unreadCount)} />
            <SummaryPill label="总数" value={String(notifications.length)} />
            <SummaryPill label="活跃筛选" value={activeFilter === 'all' ? '全部' : '未读'} />
          </div>

          <div className="mt-3 flex gap-2 overflow-x-auto pb-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
            {filters.map((filter) => (
              <button
                key={filter.key}
                onClick={() => setActiveFilter(filter.key)}
                className={`rounded-full border px-3 py-2 text-xs font-semibold whitespace-nowrap transition ${
                  activeFilter === filter.key
                    ? 'border-cyan-300/80 bg-cyan-300 text-slate-950 shadow-[0_0_24px_rgba(103,232,249,0.35)]'
                    : 'border-cyan-400/10 bg-slate-900/80 text-slate-300'
                }`}
              >
                {filter.label}
              </button>
            ))}
          </div>
        </header>

        <div className="flex-1 px-4 py-4">
          <div className="mb-3 flex items-center justify-between gap-3">
            <p className="text-[11px] uppercase tracking-[0.2em] text-slate-500">信号列表</p>
            <button
              onClick={() => void handleMarkAllRead()}
              disabled={unreadCount === 0}
              className="rounded-full border border-cyan-400/10 bg-slate-950/80 px-3 py-1.5 text-[11px] text-slate-400 disabled:opacity-40"
            >
              全部标为已读
            </button>
          </div>

          {loading ? (
            <div className="rounded-[24px] border border-cyan-400/10 bg-slate-950/80 px-4 py-10 text-center text-slate-400">
              加载通知中...
            </div>
          ) : error ? (
            <div className="rounded-[24px] border border-rose-400/20 bg-rose-500/10 px-4 py-6 text-center text-rose-200">
              <p>{error}</p>
              <button onClick={() => void loadNotifications()} className="mt-3 text-sm font-semibold text-cyan-300 underline">
                重试
              </button>
            </div>
          ) : (
            <NotificationList notifications={visibleNotifications} onOpen={(notification) => void handleOpen(notification)} />
          )}
        </div>

      </div>
    </div>
    </PullToRefresh>
  )
}

function SummaryPill({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-cyan-400/10 bg-cyan-400/5 px-3 py-3 shadow-[0_0_24px_rgba(34,211,238,0.06)]">
      <p className="text-[10px] uppercase tracking-[0.18em] text-slate-500">{label}</p>
      <p className="mt-2 text-base font-black text-cyan-200">{value}</p>
    </div>
  )
}
