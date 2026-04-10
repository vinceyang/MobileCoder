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

  const queueSummary = useMemo(() => {
    const running = tasks.filter((task) => task.state === 'running').length
    const waiting = tasks.filter((task) => task.state === 'waiting').length
    const attention = tasks.filter((task) => task.state === 'attention').length
    return {
      running,
      waiting,
      attention,
    }
  }, [tasks])

  const hotTask = visibleTasks[0] ?? tasks[0] ?? null
  const hotTaskEventKind = hotTask?.timeline?.[0]?.kind || 'info'

  const handleLogout = () => {
    localStorage.clear()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-[#020816] text-slate-100">
      <div className="mx-auto flex min-h-screen w-full max-w-md flex-col">
        <header className="border-b border-cyan-400/10 bg-slate-950/50 px-4 pt-5 pb-4 backdrop-blur">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0">
              <p className="text-[11px] uppercase tracking-[0.24em] text-cyan-300">Control Tower</p>
              <h1 className="mt-3 text-3xl font-black tracking-tight text-slate-50">Tasks</h1>
              <p className="mt-2 text-sm text-slate-400">
                {queueSummary.running} tasks active, {queueSummary.waiting} waiting,{' '}
                {queueSummary.attention} attention needed
              </p>
            </div>
            <div className="rounded-2xl border border-cyan-400/10 bg-cyan-400/5 px-3 py-2 shadow-[0_0_32px_rgba(34,211,238,0.08)]">
              <p className="text-[10px] uppercase tracking-[0.18em] text-slate-500">hot lane</p>
              <p className="mt-2 max-w-24 truncate text-base font-black text-cyan-300">
                {hotTask?.tool || 'standby'}
              </p>
              <p className="mt-1 max-w-24 truncate text-[11px] text-slate-400">
                {hotTask?.title || 'No active queue'}
              </p>
            </div>
          </div>

          <div className="mt-4 grid grid-cols-[minmax(0,1fr)_auto] gap-3">
            <div className="rounded-[24px] border border-cyan-400/10 bg-slate-900/80 px-3 py-3 shadow-[0_0_40px_rgba(8,145,178,0.12)]">
              <div className="flex items-center justify-between gap-3">
                <div>
                  <p className="text-[10px] uppercase tracking-[0.18em] text-slate-500">live signal</p>
                  <p className="mt-2 text-sm font-semibold text-slate-100">
                    {hotTask?.recent_event || hotTask?.summary || 'Queue is quiet for the moment'}
                  </p>
                </div>
                <span className={`h-2.5 w-2.5 rounded-full ${hotTask ? 'bg-cyan-300 shadow-[0_0_18px_rgba(103,232,249,0.85)]' : 'bg-slate-600'}`} />
              </div>
              <div className="mt-3 flex items-center gap-2 text-[11px] text-slate-400">
                <span className={`rounded-full px-2 py-1 ${taskEventKindStyles[hotTaskEventKind]}`}>
                  {taskEventKindLabels[hotTaskEventKind]}
                </span>
                <span className="truncate">
                  {hotTask ? formatActivityLabel(hotTask.last_activity_at) : 'Awaiting task activity'}
                </span>
              </div>
            </div>
            <div className="flex flex-col gap-2">
              <button
                onClick={() => navigate('/devices')}
                className="rounded-2xl border border-cyan-400/10 bg-slate-900/70 px-3 py-3 text-left text-[11px] uppercase tracking-[0.16em] text-cyan-300"
              >
                Devices
              </button>
              <button
                onClick={handleLogout}
                className="rounded-2xl border border-slate-800 bg-slate-950/90 px-3 py-3 text-left text-[11px] uppercase tracking-[0.16em] text-slate-400"
              >
                Logout
              </button>
            </div>
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
              <p className="text-[11px] uppercase tracking-[0.2em] text-slate-500">priority queue</p>
              <p className="mt-1 text-sm text-slate-400">Recent events lead, metadata trails.</p>
            </div>
            <span className="rounded-full border border-cyan-400/10 bg-slate-950/80 px-2.5 py-1 text-[11px] text-slate-400">
              {visibleTasks.length} visible
            </span>
          </div>

          {loading ? (
            <div className="rounded-[24px] border border-cyan-400/10 bg-slate-950/80 px-4 py-10 text-center text-slate-400">
              加载任务中...
            </div>
          ) : error ? (
            <div className="mt-6 rounded-[24px] border border-rose-400/20 bg-rose-500/10 px-4 py-6 text-center text-rose-200">
              <p>{error}</p>
              <button onClick={() => void loadTasks()} className="mt-3 text-sm font-semibold text-cyan-300 underline">
                重试
              </button>
            </div>
          ) : visibleTasks.length === 0 ? (
            <div className="mt-6 rounded-[24px] border border-cyan-400/10 bg-slate-950/80 px-4 py-8 text-center text-slate-400">
              <p>当前没有匹配的任务</p>
              <p className="mt-2 text-sm">打开设备页查看 Session，或等待 Agent 创建新任务。</p>
            </div>
          ) : (
            <div className="space-y-3">
              {visibleTasks.map((task, index) => {
                const latestEventKind = task.timeline?.[0]?.kind || 'info'
                const isHotTask = index === 0

                return (
                  <button
                    key={task.id}
                    type="button"
                    onClick={() => navigate(`/tasks/${encodeURIComponent(task.id)}`)}
                    className={`w-full rounded-[22px] border p-4 text-left transition active:scale-[0.99] ${
                      isHotTask
                        ? 'border-cyan-300/30 bg-[linear-gradient(180deg,rgba(8,145,178,0.24),rgba(2,8,22,0.92))] shadow-[0_0_44px_rgba(34,211,238,0.14)]'
                        : 'border-cyan-400/10 bg-slate-950/80'
                    }`}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <h3 className={`truncate text-base font-extrabold ${isHotTask ? 'text-cyan-50' : 'text-slate-50'}`}>
                          {task.title}
                        </h3>
                        <p className="mt-1 truncate text-xs text-slate-500">
                          {task.project_path || task.device_name}
                        </p>
                      </div>
                      <span
                        className={`rounded-full px-2 py-1 text-[11px] whitespace-nowrap ${taskStateStyles[task.state]}`}
                      >
                        {taskStateLabels[task.state]}
                      </span>
                    </div>

                    <div className="mt-3 flex items-center gap-2 flex-wrap">
                      <span className={`rounded-full px-2 py-1 text-[11px] ${taskEventKindStyles[latestEventKind]}`}>
                        {taskEventKindLabels[latestEventKind]}
                      </span>
                      <p className={`min-w-0 flex-1 text-sm ${isHotTask ? 'text-slate-100' : 'text-slate-200'}`}>
                        {task.recent_event || task.summary}
                      </p>
                    </div>

                    <div className="mt-4 flex items-center justify-between gap-3 text-[11px] text-slate-400">
                      <span className="truncate">{task.device_name}</span>
                      <span className="shrink-0">{formatActivityLabel(task.last_activity_at)}</span>
                    </div>
                  </button>
                )
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
