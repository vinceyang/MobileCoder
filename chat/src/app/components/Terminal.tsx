'use client';

import { useEffect, useRef, useState, useMemo } from 'react';
import AnsiToHtml from 'ansi-to-html';

interface TerminalProps {
  deviceId: string;
}

interface KeyButton {
  label: string;
  key: string;
  mods?: string[];
  description?: string;
}

// 获取 WebSocket 地址 - 支持手机端访问
const getWSUrl = (deviceId: string, token: string) => {
  if (typeof window === 'undefined') return `ws://localhost:8080/ws?device_id=${deviceId}&token=${token}`;
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const hostname = window.location.hostname;
  return `${protocol}//${hostname}:8080/ws?device_id=${deviceId}&token=${token}`;
};

// 检测是否为移动端
const isMobile = typeof window !== 'undefined' && window.innerWidth < 768;

export default function Terminal({ deviceId }: TerminalProps) {
  const [output, setOutput] = useState('');
  const [input, setInput] = useState('');
  const [connected, setConnected] = useState(false);
  const [mode, setMode] = useState<'text' | 'keys'>(isMobile ? 'keys' : 'text');
  const [lastKey, setLastKey] = useState<string>('');
  const [isMobileView, setIsMobileView] = useState(isMobile);
  const wsRef = useRef<WebSocket | null>(null);
  const outputRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  // ANSI to HTML 转换器
  const ansiToHtml = useMemo(() => new AnsiToHtml({
    fg: '#10b981', // 默认绿色
    bg: '#000000',
    newline: true,
    escapeXML: true,
    colors: {
      0: '#000000',
      1: '#ff0000',
      2: '#00ff00',
      3: '#ffff00',
      4: '#0000ff',
      5: '#ff00ff',
      6: '#00ffff',
      7: '#ffffff',
      8: '#808080',
      9: '#ff0000',
    }
  }), []);

  // 监听窗口大小变化
  useEffect(() => {
    const handleResize = () => {
      setIsMobileView(window.innerWidth < 768);
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    const token = localStorage.getItem('token') || 'viewer';
    const wsUrl = getWSUrl(deviceId, token);
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      setConnected(true);
      console.log('WebSocket connected');
    };

    ws.onclose = () => {
      setConnected(false);
      console.log('WebSocket disconnected');
    };

    ws.onerror = () => {
      // WebSocket 连接失败是正常的（设备未绑定时），静默处理
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      if (msg.type === 'terminal_output') {
        setOutput(msg.payload.content);
      }
    };
    wsRef.current = ws;

    return () => ws.close();
  }, [deviceId]);

  useEffect(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight;
    }
  }, [output]);

  // 发送按键到服务器
  const sendKey = (key: string, modifiers: string[] = []) => {
    if (!wsRef.current) return;

    // 检查是否是斜杠命令
    if (key.startsWith('/')) {
      // 斜杠命令作为文本发送
      wsRef.current.send(JSON.stringify({
        type: 'terminal_input',
        payload: { content: key + '\n' }
      }));
      setLastKey(key);
      setTimeout(() => setLastKey(''), 500);
      return;
    }

    wsRef.current.send(JSON.stringify({
      type: 'terminal_input',
      payload: {
        key: key,
        modifiers: modifiers,
        action: 'key'
      }
    }));
    const keyName = modifiers.length > 0 ? `${modifiers.join('+')}+${key}` : key;
    setLastKey(keyName);
    setTimeout(() => setLastKey(''), 500);
  };

  const handleSend = () => {
    if (!input.trim() || !wsRef.current) return;
    wsRef.current.send(JSON.stringify({
      type: 'terminal_input',
      payload: { content: input + '\n' }
    }));
    setInput('');
    inputRef.current?.focus();
  };

  // Claude 常用命令和快捷键分组
  const commandGroups: { name: string; keys: KeyButton[] }[] = [
    {
      name: '常用命令',
      keys: [
        { label: '/help', key: '/help', description: '帮助' },
        { label: '/clear', key: '/clear', description: '清屏' },
        { label: '/model', key: '/model', description: '切换模型' },
        { label: '/memory', key: '/memory', description: '记忆' },
      ]
    },
    {
      name: '执行控制',
      keys: [
        { label: 'Esc', key: 'Escape', description: '取消' },
        { label: '↵ 执行', key: 'Enter', description: '执行命令' },
        { label: 'Ctrl+C', key: 'c', mods: ['ctrl'], description: '中断' },
        { label: 'Ctrl+L', key: 'l', mods: ['ctrl'], description: '清屏' },
      ]
    },
    {
      name: '补全与历史',
      keys: [
        { label: '⇥ Tab', key: 'Tab', description: '补全' },
        { label: '⇤ S+Tab', key: 'Tab', mods: ['shift'], description: '反向' },
        { label: '↑', key: 'Up', description: '上条' },
        { label: '↓', key: 'Down', description: '下条' },
      ]
    },
  ];

  return (
    <div className="h-[100dvh] bg-gray-900 flex flex-col overflow-y-auto overflow-x-hidden touch-pan-y">
      {/* 头部 - 紧凑设计 */}
      <div className={`flex items-center justify-between px-2 md:px-4 py-2 border-b border-gray-800 ${isMobileView ? 'min-h-[44px]' : ''}`}>
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-400">终端</span>
          <span className={`text-xs ${connected ? 'text-green-400' : 'text-red-400'}`}>
            {connected ? '●' : '○'}
          </span>
        </div>
      </div>

      {/* 终端输出区域 */}
      <div
        ref={outputRef}
        className="flex-1 overflow-auto p-2 md:p-4"
        style={{ WebkitOverflowScrolling: 'touch' }}
      >
        {!output ? (
          <div className="text-gray-500 text-sm">
            {connected ? '等待终端输出...' : '请启动 Desktop Agent 连接设备'}
          </div>
        ) : (
          <pre
            className="whitespace-pre-wrap font-mono text-xs md:text-sm"
            dangerouslySetInnerHTML={{ __html: ansiToHtml.toHtml(output) }}
          />
        )}
      </div>

      {/* 按键反馈提示 */}
      {lastKey && (
        <div className="fixed bottom-32 left-1/2 -translate-x-1/2 bg-blue-600/90 text-white px-3 py-1.5 rounded-lg text-xs md:text-sm animate-pulse z-50">
          已发送: {lastKey}
        </div>
      )}

      {/* 底部输入/快捷键区域 */}
      {mode === 'text' ? (
        <div className="flex p-2 md:p-4 border-t border-gray-800 gap-2">
          <textarea
            ref={inputRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                handleSend();
              }
            }}
            className="flex-1 px-3 md:px-4 py-2 md:py-3 bg-gray-800 border border-gray-700 rounded-lg text-white font-mono text-sm md:text-base resize-none"
            placeholder="输入指令..."
            rows={1}
            autoFocus
          />
          <button
            onClick={handleSend}
            className="px-3 md:px-4 py-2 md:py-3 bg-blue-600 hover:bg-blue-500 active:bg-blue-700 text-white font-medium rounded-lg transition-colors text-sm"
          >
            发送
          </button>
          <button
            onClick={() => setMode('keys')}
            className="px-3 md:px-4 py-2 md:py-3 bg-gray-700 hover:bg-gray-600 active:bg-gray-500 text-white font-medium rounded-lg transition-colors text-sm"
          >
            快捷
          </button>
        </div>
      ) : (
        <div className="flex flex-col p-2 md:p-4 border-t border-gray-800 max-h-[50dvh]">
          {/* 快捷键网格 */}
          <div className="grid grid-cols-4 gap-1 md:gap-2 flex-1 overflow-y-auto">
            {commandGroups.flatMap(group =>
              group.keys.map((btn) => (
                <button
                  key={btn.label}
                  onClick={() => sendKey(btn.key, btn.mods || [])}
                  className={`flex flex-col items-center justify-center py-2 md:py-2.5 rounded-lg transition-colors active:scale-95 ${
                    btn.key.startsWith('/')
                      ? 'bg-purple-700 hover:bg-purple-600 active:bg-purple-500'
                      : 'bg-gray-700 hover:bg-gray-600 active:bg-blue-600'
                  } text-white`}
                >
                  <span className="font-mono font-bold text-xs md:text-sm">{btn.label}</span>
                  <span className="text-[10px] md:text-xs text-gray-300 mt-0.5">{btn.description}</span>
                </button>
              ))
            )}
          </div>
          {/* 切换到文本模式的按钮 */}
          <button
            onClick={() => setMode('text')}
            className="mt-2 w-full py-2 bg-gray-700 hover:bg-gray-600 active:bg-gray-500 text-white font-medium rounded-lg transition-colors text-sm"
          >
            文本
          </button>
        </div>
      )}
    </div>
  );
}
