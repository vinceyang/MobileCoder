'use client'

import type { NotificationRecord } from '@/lib/notifications'
import {
  buildNotificationTaskHref,
  formatNotificationTime,
  getNotificationEventLabel,
  getNotificationEventStyle,
  isUnreadNotification,
} from '@/lib/notifications'

interface NotificationListProps {
  notifications: NotificationRecord[]
  onOpen: (notification: NotificationRecord) => void
}

export function NotificationList({ notifications, onOpen }: NotificationListProps) {
  if (notifications.length === 0) {
    return (
      <div className="rounded-[24px] border border-cyan-400/10 bg-slate-950/80 px-4 py-8 text-center text-slate-400">
        当前没有通知
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {notifications.map((notification) => {
        const unread = isUnreadNotification(notification)
        const targetHref = buildNotificationTaskHref(notification)

        return (
          <button
            key={notification.id}
            type="button"
            onClick={() => onOpen(notification)}
            className={`w-full rounded-[24px] border px-4 py-4 text-left transition active:scale-[0.99] ${
              unread
                ? 'border-cyan-300/25 bg-[linear-gradient(180deg,rgba(8,145,178,0.16),rgba(2,8,22,0.94))] shadow-[0_0_36px_rgba(34,211,238,0.08)]'
                : 'border-cyan-400/10 bg-slate-950/80'
            }`}
          >
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <div className="flex items-center gap-2 flex-wrap">
                  <h3 className="truncate text-base font-extrabold text-slate-50">{notification.title}</h3>
                  <span
                    className={`rounded-full px-2 py-1 text-[11px] whitespace-nowrap ${
                      unread
                        ? 'border border-cyan-300/30 bg-cyan-300/10 text-cyan-100'
                        : 'border border-slate-700 bg-slate-900/80 text-slate-400'
                    }`}
                  >
                    {unread ? '未读' : '已读'}
                  </span>
                </div>
                <p className="mt-2 truncate text-xs text-slate-500">{notification.session_name || notification.device_id}</p>
              </div>
              <span
                className={`rounded-full px-2 py-1 text-[11px] whitespace-nowrap ${getNotificationEventStyle(notification.event_type)}`}
              >
                {getNotificationEventLabel(notification.event_type)}
              </span>
            </div>

            <p className="mt-3 text-sm leading-6 text-slate-200">{notification.body}</p>

            <div className="mt-4 flex items-center justify-between gap-3 text-[11px] text-slate-400">
              <span className="truncate">{targetHref}</span>
              <span className="shrink-0">{formatNotificationTime(notification.created_at)}</span>
            </div>
          </button>
        )
      })}
    </div>
  )
}
