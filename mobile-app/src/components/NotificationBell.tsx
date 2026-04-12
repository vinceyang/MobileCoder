import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  useNotificationCenter,
  notificationEventDotStyle,
  notificationEventLabel,
  notificationEventStyle,
  formatNotificationTime,
} from '../services/notifications'

export function NotificationBell() {
  const navigate = useNavigate()
  const { unreadCount, syncing, notifications } = useNotificationCenter()
  const hotNotification = notifications[0]

  return (
    <button
      type="button"
      onClick={() => navigate('/notifications')}
      className="rounded-2xl border border-cyan-400/10 bg-slate-950/80 px-3 py-2 text-left shadow-[0_0_28px_rgba(34,211,238,0.08)]"
    >
      <div className="flex items-center gap-2">
        <span className="text-[11px] uppercase tracking-[0.18em] text-cyan-300">通知</span>
        <span className="rounded-full border border-cyan-400/10 bg-cyan-400/10 px-2 py-0.5 text-[11px] text-cyan-200">
          {syncing ? '同步中' : `${unreadCount} 条未读`}
        </span>
      </div>
      <div className="mt-2 flex items-center gap-2">
        <span className={`h-2.5 w-2.5 rounded-full ${hotNotification ? notificationEventDotStyle(hotNotification.event_type) : 'bg-slate-600'}`} />
        <p className="max-w-[10rem] truncate text-[11px] text-slate-400">
          {hotNotification ? hotNotification.title : '打开通知中心'}
        </p>
      </div>
      <p className="mt-1 max-w-[10rem] truncate text-[11px] text-slate-500">
        {hotNotification ? `${notificationEventLabel(hotNotification.event_type)} · ${formatNotificationTime(hotNotification.created_at)}` : '查看最近的任务提醒'}
      </p>
    </button>
  )
}

export function NotificationRuntime() {
  const navigate = useNavigate()
  const { notifications } = useNotificationCenter()

  useEffect(() => {
    const handler = (event: Event) => {
      const customEvent = event as CustomEvent<{ taskId?: string }>
      const taskId = customEvent.detail?.taskId
      if (taskId) {
        navigate(`/tasks/${encodeURIComponent(taskId)}`)
      }
    }

    window.addEventListener('mobilecoder:notification-open', handler)
    return () => {
      window.removeEventListener('mobilecoder:notification-open', handler)
    }
  }, [navigate, notifications])

  return null
}

export { notificationEventDotStyle, notificationEventLabel, notificationEventStyle, formatNotificationTime }

