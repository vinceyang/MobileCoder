import { useEffect, useRef, useState } from 'react'
import { getWsBaseUrl } from '../config/api'

interface TerminalProps {
  deviceId: string
}

interface KeyButton {
  label: string
  key: string
  mods?: string[]
  description?: string
}

const getWSUrl = (deviceId: string, sessionName: string, token: string) => {
  const params = new URLSearchParams({ device_id: deviceId, token })
  if (sessionName) {
    params.append('session_name', sessionName)
  }
  return `${getWsBaseUrl()}/ws?${params.toString()}`
}

export default function Terminal({ deviceId }: TerminalProps) {
  const [output, setOutput] = useState('')
  const [input, setInput] = useState('')
  const [connectionState, setConnectionState] = useState<'connecting' | 'connected' | 'disconnected'>('connecting')
  const [mode, setMode] = useState<'text' | 'keys'>('keys')
  const [lastKey, setLastKey] = useState<string>('')
  const wsRef = useRef<WebSocket | null>(null)
  const outputRef = useRef<HTMLDivElement>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const reconnectAttemptsRef = useRef(0)
  const connectionAttemptRef = useRef(0)

  const sessionName = localStorage.getItem('session_name') || ''

  const connect = () => {
    const connectionAttempt = ++connectionAttemptRef.current

    // Clean up existing connection
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }

    setConnectionState('connecting')
    const token = localStorage.getItem('token') || 'viewer'
    const wsUrl = getWSUrl(deviceId, sessionName, token)
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      if (connectionAttempt !== connectionAttemptRef.current) return
      setConnectionState('connected')
      reconnectAttemptsRef.current = 0
    }

    ws.onclose = () => {
      if (connectionAttempt !== connectionAttemptRef.current) return
      setConnectionState('disconnected')
      wsRef.current = null

      // Auto reconnect with exponential backoff
      const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000)
      reconnectAttemptsRef.current++
      reconnectTimeoutRef.current = setTimeout(connect, delay)
    }

    ws.onerror = () => {
      if (connectionAttempt !== connectionAttemptRef.current) return
      setConnectionState('disconnected')
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'terminal_output') {
          setOutput(msg.payload.content)
        }
      } catch (e) {
        // Ignore non-JSON messages
      }
    }
    wsRef.current = ws
  }

  useEffect(() => {
    connect()

    return () => {
      connectionAttemptRef.current++
      if (wsRef.current) {
        wsRef.current.close()
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
    }
  }, [deviceId, sessionName])

  useEffect(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight
    }
  }, [output])

  const sendKey = (key: string, modifiers: string[] = []) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return

    if (key.startsWith('/')) {
      wsRef.current.send(JSON.stringify({
        type: 'terminal_input',
        payload: { content: key + '\n' }
      }))
      setLastKey(key)
      setTimeout(() => setLastKey(''), 500)
      return
    }

    wsRef.current.send(JSON.stringify({
      type: 'terminal_input',
      payload: {
        key: key,
        modifiers: modifiers,
        action: 'key'
      }
    }))
    const keyName = modifiers.length > 0 ? `${modifiers.join('+')}+${key}` : key
    setLastKey(keyName)
    setTimeout(() => setLastKey(''), 500)
  }

  const handleSend = () => {
    if (!input.trim() || !wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) return
    wsRef.current.send(JSON.stringify({
      type: 'terminal_input',
      payload: { content: input + '\n' }
    }))
    setInput('')
  }

  const commandGroups: { name: string; keys: KeyButton[] }[] = [
    {
      name: '常用命令',
      keys: [
        { label: '/help', key: '/help', description: '帮助' },
        { label: '/clear', key: '/clear', description: '清屏' },
      ]
    },
    {
      name: '执行控制',
      keys: [
        { label: 'Esc', key: 'Escape', description: '取消' },
        { label: '↵', key: 'Enter', description: '执行' },
        { label: 'Ctrl+C', key: 'c', mods: ['ctrl'], description: '中断' },
      ]
    },
    {
      name: '导航',
      keys: [
        { label: '↑', key: 'Up', description: '上条' },
        { label: '↓', key: 'Down', description: '下条' },
        { label: 'Tab', key: 'Tab', description: '补全' },
      ]
    },
  ]

  return (
    <div className="h-full bg-gray-900 flex flex-col overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-800">
        <div className="flex items-center gap-2">
          <span
            className={`w-2 h-2 rounded-full ${
              connectionState === 'connected'
                ? 'bg-emerald-400 shadow-[0_0_12px_rgba(52,211,153,0.45)]'
                : connectionState === 'connecting'
                  ? 'bg-amber-400 shadow-[0_0_12px_rgba(251,191,36,0.35)]'
                  : 'bg-rose-400 shadow-[0_0_12px_rgba(251,113,133,0.35)]'
            }`}
          />
          <span className="text-xs text-gray-400">
            {connectionState === 'connected'
              ? '已连接'
              : connectionState === 'connecting'
                ? '连接中...'
                : '连接已断开'}
          </span>
        </div>
        <button
          onClick={connect}
          className="text-xs text-cyan-300 hover:text-cyan-200"
        >
          重连
        </button>
      </div>

      <div ref={outputRef} className="flex-1 overflow-auto p-4">
        {!output ? (
          <div className="text-gray-500 text-sm">
            {connectionState === 'connected'
              ? '等待终端输出...'
              : connectionState === 'connecting'
                ? '正在建立终端连接...'
                : '请启动 Desktop Agent 或点击重连'}
          </div>
        ) : (
          <pre className="whitespace-pre-wrap font-mono text-xs md:text-sm text-green-400">
            {output}
          </pre>
        )}
      </div>

      {lastKey && (
        <div className="fixed bottom-32 left-1/2 z-50 -translate-x-1/2 rounded-lg bg-cyan-400/90 px-3 py-1.5 text-xs text-slate-950 animate-pulse">
          已发送: {lastKey}
        </div>
      )}

      {mode === 'text' ? (
        <div className="flex p-4 border-t border-gray-800 gap-2">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                handleSend()
              }
            }}
            className="flex-1 px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg text-white font-mono text-sm resize-none"
            placeholder="输入指令..."
            rows={1}
          />
          <button
            onClick={handleSend}
            className="rounded-lg bg-cyan-400 px-4 py-3 font-medium text-slate-950 hover:bg-cyan-300"
          >
            发送
          </button>
          <button
            onClick={() => setMode('keys')}
            className="px-4 py-3 bg-gray-700 hover:bg-gray-600 text-white rounded-lg"
          >
            快捷
          </button>
        </div>
      ) : (
        <div className="flex flex-col p-4 border-t border-gray-800 max-h-[50dvh]">
          <div className="grid grid-cols-4 gap-2 flex-1 overflow-y-auto">
            {commandGroups.flatMap(group =>
              group.keys.map((btn) => (
                <button
                  key={btn.label}
                  onClick={() => sendKey(btn.key, btn.mods || [])}
                  className={`flex flex-col items-center justify-center py-2 rounded-lg active:scale-95 ${
                    btn.key.startsWith('/')
                      ? 'bg-[linear-gradient(180deg,rgba(34,211,238,0.28),rgba(8,47,73,0.98))] border border-cyan-300/25 hover:border-cyan-200/40 hover:bg-[linear-gradient(180deg,rgba(34,211,238,0.36),rgba(8,47,73,0.98))]'
                      : 'bg-slate-700 hover:bg-slate-600 active:bg-cyan-700'
                  } text-white`}
                >
                  <span className="font-mono font-bold text-sm">{btn.label}</span>
                  <span className="text-[10px] text-gray-300 mt-0.5">{btn.description}</span>
                </button>
              ))
            )}
          </div>
          <button
            onClick={() => setMode('text')}
            className="mt-2 w-full rounded-lg bg-slate-700 py-2 text-white hover:bg-slate-600"
          >
            文本输入
          </button>
        </div>
      )}
    </div>
  )
}
