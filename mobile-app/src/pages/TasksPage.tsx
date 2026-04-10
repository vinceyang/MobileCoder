import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  formatActivityLabel,
  getTasks,
  Task,
  TaskState,
  taskEventKindLabels,
  taskEventKindStyles,
  taskStateLabels,
  taskStateStyles,
} from '../services/tasks'

const filters: Array<{ key: 'all' | TaskState; label: string }> = [
  { key: 'all', label: '全部' },
  { key: 'running', label: '运行中' },
  { key: 'waiting', label: '等待输入' },
  { key: 'attention', label: '需关注' },
  { key: 'completed', label: '已完成' },
]

export default function TasksPage() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [activeFilter, setActiveFilter] = useState<'all' | TaskState>('all')
  const navigate = useNavigate()

  useEffect(() => {
    void loadTasks()
  }, [])

  async function loadTasks() {
    try {
      setError('')
      const data = await getTasks()
      setTasks(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch tasks')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const visibleTasks = useMemo(() => {
    if (activeFilter === 'all') return tasks
    return tasks.filter((task) => task.state === activeFilter)
  }, [activeFilter, tasks])

  const handleLogout = () => {
    localStorage.clear()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-gray-900">
      <header className="p-4 border-b border-gray-800">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-xs text-sky-300 uppercase tracking-[0.2em]">Control Tower</p>
            <h1 className="text-2xl font-bold mt-2">Tasks</h1>
            <p className="text-sm text-gray-400 mt-2">先看任务，再决定是否进入终端。</p>
          </div>
          <button onClick={handleLogout} className="text-gray-400 text-sm">
            退出
          </button>
        </div>

        <div className="flex gap-2 overflow-x-auto mt-4 pb-1">
          {filters.map((filter) => (
            <button
              key={filter.key}
              onClick={() => setActiveFilter(filter.key)}
              className={`px-3 py-2 rounded-full text-sm whitespace-nowrap border ${
                activeFilter === filter.key
                  ? 'bg-sky-400 text-slate-950 border-sky-300'
                  : 'bg-gray-800 text-gray-300 border-gray-700'
              }`}
            >
              {filter.label}
            </button>
          ))}
        </div>

        <button
          onClick={() => navigate('/devices')}
          className="mt-4 text-sm text-sky-300"
        >
          查看设备 →
        </button>
      </header>

      <div className="p-4">
        {loading ? (
          <p className="text-gray-400 text-center">加载任务中...</p>
        ) : error ? (
          <div className="text-center text-red-400 mt-8">
            <p>{error}</p>
            <button onClick={() => void loadTasks()} className="mt-2 text-blue-400 underline">
              重试
            </button>
          </div>
        ) : visibleTasks.length === 0 ? (
          <div className="text-center text-gray-400 mt-8">
            <p>当前没有匹配的任务</p>
            <p className="text-sm mt-2">打开设备页查看 Session，或等待 Agent 创建新任务。</p>
          </div>
        ) : (
          <div className="space-y-3">
            {visibleTasks.map((task) => {
              const latestEventKind = task.timeline?.[0]?.kind || 'info'
              return (
                <button
                  key={task.id}
                  type="button"
                  onClick={() => navigate(`/tasks/${encodeURIComponent(task.id)}`)}
                  className="w-full text-left bg-gray-800 p-4 rounded-2xl active:bg-gray-700 border border-gray-700"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <h3 className="font-medium text-white truncate">{task.title}</h3>
                        <span className="text-[11px] uppercase tracking-wide text-sky-300 bg-sky-500/10 px-2 py-1 rounded-full border border-sky-500/20">
                          {task.tool}
                        </span>
                      </div>
                      <p className="text-sm text-gray-400 mt-1 truncate">
                        {task.project_path || task.device_name}
                      </p>
                    </div>
                    <span className={`px-2 py-1 rounded-full text-xs whitespace-nowrap ${taskStateStyles[task.state]}`}>
                      {taskStateLabels[task.state]}
                    </span>
                  </div>

                  <div className="mt-4 flex items-center gap-2 flex-wrap">
                    <span className={`text-[11px] px-2 py-1 rounded-full ${taskEventKindStyles[latestEventKind]}`}>
                      {taskEventKindLabels[latestEventKind]}
                    </span>
                    <p className="text-sm text-gray-200 min-w-0 flex-1">{task.recent_event || task.summary}</p>
                  </div>

                  <div className="mt-4 flex items-center justify-between text-xs text-gray-400">
                    <span>{task.device_name}</span>
                    <span>{formatActivityLabel(task.last_activity_at)}</span>
                  </div>
                </button>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
