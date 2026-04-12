'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { Bell } from 'lucide-react'
import { getUnreadNotificationCount } from '@/lib/notifications'

interface NotificationBellProps {
  className?: string
}

export function NotificationBell({ className = '' }: NotificationBellProps) {
  const [count, setCount] = useState(0)

  useEffect(() => {
    let active = true

    async function refresh() {
      try {
        const unreadCount = await getUnreadNotificationCount()
        if (active) {
          setCount(unreadCount)
        }
      } catch {
        if (active) {
          setCount(0)
        }
      }
    }

    void refresh()
    const interval = window.setInterval(() => {
      void refresh()
    }, 30000)

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        void refresh()
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)

    return () => {
      active = false
      window.clearInterval(interval)
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [])

  return (
    <Link
      href="/notifications"
      className={`relative inline-flex items-center gap-2 rounded-2xl border border-cyan-400/10 bg-slate-950/80 px-3 py-2 text-sm text-slate-200 shadow-[0_0_24px_rgba(34,211,238,0.08)] transition hover:border-cyan-300/30 hover:text-cyan-100 ${className}`}
    >
      <Bell className="h-4 w-4" />
      <span className="hidden sm:inline">通知</span>
      {count > 0 ? (
        <span className="ml-1 inline-flex min-w-5 items-center justify-center rounded-full bg-cyan-300 px-1.5 py-0.5 text-[10px] font-black text-slate-950">
          {count > 99 ? '99+' : count}
        </span>
      ) : (
        <span className="ml-1 inline-flex min-w-5 items-center justify-center rounded-full border border-slate-700 bg-slate-900/90 px-1.5 py-0.5 text-[10px] font-black text-slate-500">
          0
        </span>
      )}
    </Link>
  )
}
