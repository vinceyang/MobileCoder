import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  formatActivityLabel,
  getTask,
  Task,
  taskEventKindLabels,
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

  return (
    <div className="min-h-screen bg-gray-900">
      <header className="p-4 border-b border-gray-800">
        <button onClick={() => navigate('/tasks')} className="text-gray-400 mb-4">
          ← 返回任务列表
        </button>
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-xs text-sky-300 uppercase tracking-[0.2em]">Task Detail</p>
            <h1 className="text-2xl font-bold mt-2">{task.title}</h1>
            <p className="text-sm text-gray-400 mt-2">{task.summary}</p>
          </div>
          <span className={`px-2 py-1 rounded-full text-xs whitespace-nowrap ${taskStateStyles[task.state]}`}>
            {taskStateLabels[task.state]}
          </span>
        </div>
      </header>

      <div className="p-4 space-y-4">
        <InfoCard label="状态原因" value={task.state_reason} />
        <InfoCard label="最近事件" value={task.recent_event} />
        <InfoCard label="项目路径" value={task.project_path || '未提供'} />
        <InfoCard label="设备" value={task.device_name} />
        <InfoCard label="Session" value={task.session_name} />
        <InfoCard label="最近活动" value={formatActivityLabel(task.last_activity_at)} />

        <div className="rounded-2xl border border-gray-700 bg-gray-800 p-4">
          <p className="text-xs uppercase tracking-wide text-gray-500">Recent Timeline</p>
          <p className="text-sm text-gray-400 mt-2">来自实时 terminal output 的最近事件摘录。</p>

          {task.timeline && task.timeline.length > 0 ? (
            <div className="mt-4 space-y-3">
              {task.timeline.map((event, index) => (
                <div key={`${event.timestamp}-${index}`} className="flex gap-3">
                  <div className="mt-1 h-2.5 w-2.5 rounded-full bg-sky-400 shrink-0" />
                  <div className="min-w-0">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className={`text-[11px] px-2 py-1 rounded-full ${taskEventKindStyles[event.kind]}`}>
                        {taskEventKindLabels[event.kind]}
                      </span>
                      <p className="text-sm text-gray-100 break-words">{event.summary}</p>
                    </div>
                    <p className="text-xs text-gray-500 mt-1">{formatActivityLabel(event.timestamp)}</p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="mt-4 rounded-2xl border border-dashed border-gray-700 p-4 text-sm text-gray-400">
              还没有捕获到实时事件，当前仍回退使用状态摘要。
            </div>
          )}
        </div>

        <div className="flex gap-3 pt-2">
          <button
            onClick={openTerminal}
            className="flex-1 px-4 py-3 rounded-2xl bg-sky-400 text-slate-950 font-medium"
          >
            打开终端接管
          </button>
          <button
            onClick={() => void loadTask(task.id)}
            className="px-4 py-3 rounded-2xl bg-gray-700 text-gray-100"
          >
            刷新
          </button>
        </div>
      </div>
    </div>
  )
}

function InfoCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-gray-700 bg-gray-800 p-4">
      <p className="text-xs uppercase tracking-wide text-gray-500">{label}</p>
      <p className="text-sm text-gray-100 mt-2 break-all">{value}</p>
    </div>
  )
}
