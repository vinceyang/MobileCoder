import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  formatActivityLabel,
  getTask,
  Task,
  taskEventKindLabels,
  taskEventKindDotStyles,
  taskEventKindStyles,
  taskStateLabels,
  taskStateStyles,
} from '../services/tasks'

export default function TaskDetailPage() {
  const { taskId } = useParams<{ taskId: string }>()
  const [task, setTask] = useState<Task | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const navigate = useNavigate()

  useEffect(() => {
    if (!taskId) {
      setError('Task not found')
      setLoading(false)
      return
    }

    void loadTask(decodeURIComponent(taskId))
  }, [taskId])

  async function loadTask(id: string) {
    try {
      setLoading(true)
      setError('')
      const data = await getTask(id)
      setTask(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch task')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const openTerminal = () => {
    if (!task) return
    localStorage.setItem('device_id', task.device_id)
    localStorage.setItem('session_name', task.session_name)
    navigate(`/terminal?device_id=${task.device_id}&session_name=${encodeURIComponent(task.session_name)}&task_id=${encodeURIComponent(task.id)}`)
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-900 flex items-center justify-center">
        <p className="text-gray-400">加载任务中...</p>
      </div>
    )
  }

  if (error || !task) {
    return (
      <div className="min-h-screen bg-gray-900 p-4">
        <button onClick={() => navigate('/tasks')} className="text-gray-400 mb-4">
          ← 返回任务列表
        </button>
        <div className="bg-rose-500/10 border border-rose-500/20 text-rose-200 rounded-2xl p-4">
          {error || 'Task not found'}
        </div>
      </div>
    )
  }

  const metadataItems: Array<{ label: string; value: string; valueClassName?: string }> = [
    { label: '最近事件', value: task.recent_event },
    { label: '项目路径', value: task.project_path || '未提供', valueClassName: 'break-all' },
    { label: '设备', value: task.device_name },
    { label: '会话', value: task.session_name },
    { label: '最近活动', value: formatActivityLabel(task.last_activity_at) },
  ]

  return (
    <div className="min-h-screen bg-[#020816] text-slate-100">
      <div className="mx-auto flex min-h-screen w-full max-w-md flex-col">
        <header className="border-b border-cyan-400/10 bg-slate-950/50 px-4 pt-5 pb-4 backdrop-blur">
          <button onClick={() => navigate('/tasks')} className="mb-4 text-slate-400">
            ← 返回任务列表
          </button>
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <p className="text-[11px] uppercase tracking-[0.24em] text-cyan-300">任务详情</p>
              <h1 className="mt-3 text-3xl font-black tracking-tight text-slate-50">{task.title}</h1>
              <p className="mt-2 text-sm text-slate-400">{task.state_reason}</p>
            </div>
            <span
              className={`whitespace-nowrap rounded-full px-2 py-1 text-[11px] ${taskStateStyles[task.state]}`}
            >
              {taskStateLabels[task.state]}
            </span>
          </div>

          <div className="mt-4 rounded-[24px] border border-cyan-400/10 bg-[linear-gradient(180deg,rgba(8,145,178,0.18),rgba(2,8,22,0.92))] px-4 py-4 shadow-[0_0_44px_rgba(34,211,238,0.12)]">
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <p className="text-[10px] uppercase tracking-[0.18em] text-slate-500">任务概览</p>
                <p className="mt-2 text-sm font-semibold text-slate-100">{task.summary}</p>
              </div>
              <span className="mt-1 h-2.5 w-2.5 shrink-0 rounded-full bg-cyan-300 shadow-[0_0_18px_rgba(103,232,249,0.85)]" />
            </div>
            <div className="mt-4 grid grid-cols-2 gap-3">
              {metadataItems.map((item) => (
                <div key={item.label} className="rounded-2xl border border-cyan-400/10 bg-slate-950/80 px-3 py-3">
                  <p className="text-[10px] uppercase tracking-[0.18em] text-slate-500">{item.label}</p>
                  <p className={`mt-2 text-sm text-slate-100 ${item.valueClassName || ''}`}>{item.value}</p>
                </div>
              ))}
            </div>
          </div>
        </header>

        <div className="flex-1 space-y-4 px-4 py-4">
          <div className="rounded-[28px] border border-cyan-400/15 bg-[linear-gradient(180deg,rgba(15,23,42,0.96),rgba(2,8,22,0.98))] p-4 shadow-[0_0_52px_rgba(34,211,238,0.12)]">
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="text-[11px] uppercase tracking-[0.22em] text-cyan-300">实时时间线</p>
                <p className="mt-2 text-sm text-slate-400">实时 terminal output 的关键节点，优先用于判断是否立即接管。</p>
              </div>
              <span className="rounded-full border border-cyan-400/10 bg-cyan-400/5 px-2.5 py-1 text-[11px] text-slate-400">
                {task.timeline?.length || 0} 条事件
              </span>
            </div>

            {task.timeline && task.timeline.length > 0 ? (
              <div className="mt-5 space-y-4">
                {task.timeline.map((event, index) => (
                  <div
                    key={`${event.timestamp}-${index}`}
                    className="grid grid-cols-[12px_1fr] gap-3 rounded-2xl border border-cyan-400/10 bg-slate-950/70 px-3 py-3"
                  >
                    <div className={`mt-1 h-2.5 w-2.5 rounded-full ${taskEventKindDotStyles[event.kind]}`} />
                    <div className="min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className={`text-[11px] px-2 py-1 rounded-full ${taskEventKindStyles[event.kind]}`}>
                          {taskEventKindLabels[event.kind]}
                        </span>
                        <p className="text-sm text-slate-100 break-words">{event.summary}</p>
                      </div>
                      <p className="mt-1 text-[11px] text-slate-500">{formatActivityLabel(event.timestamp)}</p>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="mt-5 rounded-2xl border border-dashed border-cyan-400/10 bg-slate-950/70 p-4 text-sm text-slate-400">
                还没有捕获到实时事件，当前仍回退使用状态摘要。
              </div>
            )}
          </div>

          <div className="flex gap-3 pt-2">
            <button
              onClick={openTerminal}
              className="flex-1 rounded-2xl bg-gradient-to-br from-cyan-300 to-sky-400 px-4 py-3 font-extrabold text-sky-950 shadow-[0_0_32px_rgba(56,189,248,0.28)]"
            >
              打开终端接管
            </button>
            <button
              onClick={() => void loadTask(task.id)}
              className="rounded-2xl border border-cyan-400/10 bg-slate-900/80 px-4 py-3 font-semibold text-slate-200"
            >
              刷新
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
