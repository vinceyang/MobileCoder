import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { NotificationBell } from '../components/NotificationBell'
import {
  formatNotificationTime,
  notificationEventDotStyle,
  notificationEventLabel,
  notificationEventStyle,
  type NotificationItem,
  useNotificationCenter,
} from '../services/notifications'

const filters: Array<{ key: 'all' | 'unread'; label: string }> = [
  { key: 'all', label: '全部' },
  { key: 'unread', label: '未读' },
]

export default function NotificationsPage() {
  const navigate = useNavigate()
  const [activeFilter, setActiveFilter] = useState<'all' | 'unread'>('all')
  const {
    notifications,
    unreadCount,
    syncing,
    error,
    refresh,
    markNotificationRead,
    markAllNotificationsRead,
  } = useNotificationCenter()

  const visibleNotifications = useMemo(() => {
    if (activeFilter === 'unread') {
      return notifications.filter((item) => !item.read_at)
    }
    return notifications
  }, [activeFilter, notifications])

  async function openNotification(notification: NotificationItem) {
    if (!notification.read_at) {
      try {
        await markNotificationRead(notification.id)
      } catch (err) {
        console.error(err)
      }
    }

    if (notification.task_id) {
      navigate(`/tasks/${encodeURIComponent(notification.task_id)}`)
      return
    }

    navigate('/tasks')
  }

  return (
    <div className="min-h-screen bg-[#020816] text-slate-100">
      <div className="mx-auto flex min-h-screen w-full max-w-md flex-col">
        <header className="border-b border-cyan-400/10 bg-slate-950/50 px-4 pt-5 pb-4 backdrop-blur">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0">
              <p className="text-[11px] uppercase tracking-[0.24em] text-cyan-300">控制塔台</p>
              <h1 className="mt-3 text-3xl font-black tracking-tight text-slate-50">通知中心</h1>
              <p className="mt-2 text-sm text-slate-400">离开电脑时，优先看这里的关键提醒。</p>
            </div>
            <NotificationBell />
          </div>

          <div className="mt-4 grid grid-cols-3 gap-2">
            <SummaryPill label="未读" value={String(unreadCount)} />
            <SummaryPill label="总数" value={String(notifications.length)} />
            <SummaryPill label="同步" value={syncing ? '进行中' : '已连接'} />
          </div>

          <div className="mt-4 flex gap-2 overflow-x-auto pb-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
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
            <div>
              <p className="text-[11px] uppercase tracking-[0.2em] text-slate-500">信号队列</p>
              <p className="mt-1 text-sm text-slate-400">点击通知后直接进入对应任务详情。</p>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => void refresh()}
                className="rounded-full border border-cyan-400/10 bg-slate-950/80 px-3 py-1.5 text-[11px] text-slate-400"
              >
                刷新
              </button>
              <button
                onClick={() => void markAllNotificationsRead()}
                disabled={unreadCount === 0}
                className="rounded-full border border-cyan-400/10 bg-slate-950/80 px-3 py-1.5 text-[11px] text-slate-400 disabled:opacity-40"
              >
                全部已读
              </button>
            </div>
          </div>

          {error ? (
            <div className="rounded-[24px] border border-rose-400/20 bg-rose-500/10 px-4 py-6 text-center text-rose-200">
              <p>{error}</p>
            </div>
          ) : visibleNotifications.length === 0 ? (
            <div className="rounded-[24px] border border-cyan-400/10 bg-slate-950/80 px-4 py-8 text-center text-slate-400">
              当前没有通知
            </div>
          ) : (
            <div className="space-y-3">
              {visibleNotifications.map((notification) => {
                const unread = !notification.read_at

                return (
                  <button
                    key={notification.id}
                    type="button"
                    onClick={() => void openNotification(notification)}
                    className={`w-full rounded-[24px] border px-4 py-4 text-left transition active:scale-[0.99] ${
                      unread
                        ? 'border-cyan-300/25 bg-[linear-gradient(180deg,rgba(8,145,178,0.16),rgba(2,8,22,0.94))] shadow-[0_0_36px_rgba(34,211,238,0.08)]'
                        : 'border-cyan-400/10 bg-slate-950/80'
                    }`}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <span className={`h-2.5 w-2.5 rounded-full ${notificationEventDotStyle(notification.event_type)}`} />
                          <h3 className="truncate text-base font-extrabold text-slate-50">{notification.title}</h3>
                        </div>
                        <p className="mt-2 truncate text-xs text-slate-500">
                          {notification.session_name || notification.device_id}
                        </p>
                      </div>
                      <span className={`rounded-full px-2 py-1 text-[11px] ${notificationEventStyle(notification.event_type)}`}>
                        {notificationEventLabel(notification.event_type)}
                      </span>
                    </div>

                    <p className="mt-3 text-sm leading-6 text-slate-200">{notification.body}</p>

                    <div className="mt-4 flex items-center justify-between gap-3 text-[11px] text-slate-400">
                      <span>{unread ? '未读' : '已读'}</span>
                      <span className="shrink-0">{formatNotificationTime(notification.created_at)}</span>
                    </div>
                  </button>
                )
              })}
            </div>
          )}
        </div>

        <footer className="border-t border-cyan-400/10 bg-slate-950/50 px-4 py-4 backdrop-blur">
          <button
            onClick={() => navigate('/tasks')}
            className="w-full rounded-2xl border border-cyan-400/10 bg-slate-900/80 px-4 py-3 text-sm font-semibold text-slate-200"
          >
            返回任务
          </button>
        </footer>
      </div>
    </div>
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
