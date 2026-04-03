import { useEffect, useRef, useState } from 'react'

interface TerminalProps {
  deviceId: string
}

interface KeyButton {
  label: string
  key: string
  mods?: string[]
  description?: string
}

const WS_SERVER = 'ws://121.41.69.142:8080'

const getWSUrl = (deviceId: string, sessionName: string, token: string) => {
  const params = new URLSearchParams({ device_id: deviceId, token })
  if (sessionName) {
    params.append('session_name', sessionName)
  }
  return `${WS_SERVER}/ws?${params.toString()}`
}

export default function Terminal({ deviceId }: TerminalProps) {
  const [output, setOutput] = useState('')
  const [input, setInput] = useState('')
  const [connected, setConnected] = useState(false)
  const [mode, setMode] = useState<'text' | 'keys'>('keys')
  const [lastKey, setLastKey] = useState<string>('')
  const wsRef = useRef<WebSocket | null>(null)
  const outputRef = useRef<HTMLDivElement>(null)

  const sessionName = localStorage.getItem('session_name') || ''

  useEffect(() => {
    const token = localStorage.getItem('token') || 'viewer'
    const wsUrl = getWSUrl(deviceId, sessionName, token)
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      setConnected(true)
    }

    ws.onclose = () => {
      setConnected(false)
    }

    ws.onerror = () => {
      // Ignore errors when device not connected
    }

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data)
      if (msg.type === 'terminal_output') {
        setOutput(msg.payload.content)
      }
    }
    wsRef.current = ws

    return () => ws.close()
  }, [deviceId, sessionName])

  useEffect(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight
    }
  }, [output])

  const sendKey = (key: string, modifiers: string[] = []) => {
    if (!wsRef.current) return

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
    if (!input.trim() || !wsRef.current) return
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
    <div className="h-[100dvh] bg-gray-900 flex flex-col overflow-y-auto touch-pan-y">
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-800">
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-400">终端</span>
          <span className={`text-xs ${connected ? 'text-green-400' : 'text-red-400'}`}>
            {connected ? '●' : '○'}
          </span>
        </div>
        <span className="text-xs text-gray-500">{sessionName}</span>
      </div>

      <div ref={outputRef} className="flex-1 overflow-auto p-4">
        {!output ? (
          <div className="text-gray-500 text-sm">
            {connected ? '等待终端输出...' : '请启动 Desktop Agent 连接设备'}
          </div>
        ) : (
          <pre className="whitespace-pre-wrap font-mono text-xs md:text-sm text-green-400">
            {output}
          </pre>
        )}
      </div>

      {lastKey && (
        <div className="fixed bottom-32 left-1/2 -translate-x-1/2 bg-blue-600/90 text-white px-3 py-1.5 rounded-lg text-xs animate-pulse z-50">
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
            className="px-4 py-3 bg-blue-600 hover:bg-blue-500 text-white rounded-lg font-medium"
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
                      ? 'bg-purple-700 hover:bg-purple-600 active:bg-purple-500'
                      : 'bg-gray-700 hover:bg-gray-600 active:bg-blue-600'
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
            className="mt-2 w-full py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg"
          >
            文本输入
          </button>
        </div>
      )}
    </div>
  )
}
