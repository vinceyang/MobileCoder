'use client';

import { useCallback, useEffect, useRef, useState, useMemo } from 'react';
import AnsiToHtml from 'ansi-to-html';
import { getWsBaseUrl } from '@/lib/api';

interface TerminalProps {
  deviceId: string;
  sessionName?: string;
  onConnectionChange?: (connected: boolean) => void;
}

interface KeyButton {
  label: string;
  key: string;
  mods?: string[];
  description?: string;
}

// 获取 WebSocket 地址 - 支持 session 路由
const getWSUrl = (deviceId: string, sessionName: string, token: string) => {
  const params = new URLSearchParams({ device_id: deviceId, token });
  if (sessionName) {
    params.append('session_name', sessionName);
  }
  return `${getWsBaseUrl()}/ws?${params.toString()}`;
};

export default function Terminal({ deviceId, sessionName, onConnectionChange }: TerminalProps) {
  const [output, setOutput] = useState('');
  const [input, setInput] = useState('');
  const [connected, setConnected] = useState(false);
  const [mode, setMode] = useState<'text' | 'keys'>('text');
  const [lastKey, setLastKey] = useState<string>('');
  const [showScrollTop, setShowScrollTop] = useState(false);
  const [showScrollBottom, setShowScrollBottom] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const outputRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);

  // 从 URL 获取 session_name
  const [urlSessionName, setUrlSessionName] = useState(() => {
    if (typeof window === 'undefined') return '';
    return new URLSearchParams(window.location.search).get('session_name') || '';
  });
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const params = new URLSearchParams(window.location.search);
      const s = params.get('session_name');
      if (!s) return;
      setUrlSessionName(s);
      localStorage.setItem('session_name', s);
    }
  }, []);

  // 优先使用 props，其次使用 URL 中的 session_name
  const storedSessionName = typeof window !== 'undefined' ? localStorage.getItem('session_name') || '' : '';
  const effectiveSessionName = sessionName || urlSessionName || storedSessionName;

  // 连接到 WebSocket
  const connect = useCallback(() => {
    if (!effectiveSessionName) return;

    // 清理现有连接
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }

    const token = localStorage.getItem('token') || 'viewer';
    const wsUrl = getWSUrl(deviceId, effectiveSessionName, token);
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      setConnected(true);
      onConnectionChange?.(true);
      reconnectAttemptsRef.current = 0;
      console.log('WebSocket connected');
    };

    ws.onclose = () => {
      setConnected(false);
      onConnectionChange?.(false);
      wsRef.current = null;

      // 自动重连，指数退避
      const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);
      reconnectAttemptsRef.current++;
      reconnectTimeoutRef.current = setTimeout(connect, delay);
      console.log(`WebSocket disconnected, reconnecting in ${delay}ms...`);
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
  }, [deviceId, effectiveSessionName, onConnectionChange]);

  // 初始化 WebSocket 连接
  useEffect(() => {
    connect();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [connect]);

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

  // 智能滚动：只在用户已经在底部时自动滚动
  useEffect(() => {
    if (outputRef.current) {
      const el = outputRef.current;
      // 检查是否接近底部（50px 范围内）
      const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 50;

      if (atBottom) {
        el.scrollTop = el.scrollHeight;
      }
      setShowScrollTop(el.scrollTop > 100);
      setShowScrollBottom(!atBottom);
    }
  }, [output]);

  const handleScroll = () => {
    if (outputRef.current) {
      const el = outputRef.current;
      const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 50;
      setShowScrollTop(el.scrollTop > 100);
      setShowScrollBottom(!atBottom);
    }
  };

  const scrollToTop = () => {
    if (outputRef.current) {
      outputRef.current.scrollTo({ top: 0, behavior: 'smooth' });
    }
  };

  const scrollToBottom = () => {
    if (outputRef.current) {
      outputRef.current.scrollTo({ top: outputRef.current.scrollHeight, behavior: 'smooth' });
    }
  };

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
    <div className="relative flex h-full flex-col overflow-hidden bg-[#020816]">
      {/* 终端输出区域 */}
      <div
        ref={outputRef}
        onScroll={handleScroll}
        className="min-h-0 flex-1 overflow-auto px-3 py-3 md:px-4"
        style={{ WebkitOverflowScrolling: 'touch' }}
      >
        {!output ? (
          <div className="rounded-2xl border border-cyan-400/10 bg-slate-950/70 p-4 text-sm text-slate-500">
            {connected ? '等待终端输出...' : '请启动 Desktop Agent 连接设备'}
          </div>
        ) : (
          <pre
            className="min-w-max whitespace-pre font-mono text-[11px] leading-5 text-emerald-200 md:text-sm"
            dangerouslySetInnerHTML={{ __html: ansiToHtml.toHtml(output) }}
          />
        )}
      </div>

      {/* 滚动按钮 */}
      {showScrollTop && (
        <button
          onClick={scrollToTop}
          className="absolute right-4 bottom-24 flex h-10 w-10 items-center justify-center rounded-full text-lg"
          style={{ backgroundColor: 'rgba(59, 130, 246, 0.9)', color: '#ffffff' }}
        >
          ↑
        </button>
      )}
      {showScrollBottom && (
        <button
          onClick={scrollToBottom}
          className="absolute right-4 bottom-40 flex h-10 w-10 items-center justify-center rounded-full text-lg"
          style={{ backgroundColor: 'rgba(59, 130, 246, 0.9)', color: '#ffffff' }}
        >
          ↓
        </button>
      )}

      {/* 按键反馈提示 */}
      {lastKey && (
        <div className="fixed bottom-24 left-1/2 z-50 -translate-x-1/2 animate-pulse rounded-lg bg-cyan-500/90 px-3 py-1.5 text-xs text-slate-950 md:text-sm">
          已发送: {lastKey}
        </div>
      )}

      {/* 底部输入/快捷键区域 */}
      {mode === 'text' ? (
        <div className="border-t border-cyan-400/10 bg-slate-950/95 p-2 backdrop-blur md:p-4">
          <div className="flex gap-2">
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
              className="min-h-[42px] flex-1 resize-none rounded-2xl border border-cyan-400/10 bg-slate-900 px-3 py-2 font-mono text-sm text-white outline-none placeholder:text-slate-600 focus:border-cyan-300/50 md:px-4 md:py-3 md:text-base"
            placeholder="输入指令..."
            rows={1}
            autoFocus
          />
          <button
            onClick={handleSend}
              className="rounded-2xl bg-cyan-300 px-4 py-2 text-sm font-black text-slate-950 transition active:scale-95 md:px-5 md:py-3"
          >
            发送
          </button>
          <button
            onClick={() => setMode('keys')}
              className="rounded-2xl border border-cyan-400/10 bg-slate-900 px-3 py-2 text-sm font-semibold text-slate-200 transition active:scale-95 md:px-4 md:py-3"
          >
            快捷
          </button>
          </div>
        </div>
      ) : (
        <div className="flex max-h-[34dvh] flex-col border-t border-cyan-400/10 bg-slate-950/95 p-2 backdrop-blur md:p-4">
          {/* 快捷键网格 */}
          <div className="grid flex-1 grid-cols-4 gap-1 overflow-y-auto md:gap-2">
            {commandGroups.flatMap(group =>
              group.keys.map((btn) => (
                <button
                  key={btn.label}
                  onClick={() => sendKey(btn.key, btn.mods || [])}
                  className={`flex flex-col items-center justify-center rounded-xl py-2 transition active:scale-95 md:py-2.5 ${
                    btn.key.startsWith('/')
                      ? 'border border-cyan-400/10 bg-cyan-500/15 text-cyan-100'
                      : 'border border-slate-700 bg-slate-800 text-white'
                  }`}
                >
                  <span className="font-mono font-bold text-xs md:text-sm">{btn.label}</span>
                  <span className="mt-0.5 text-[10px] text-slate-400 md:text-xs">{btn.description}</span>
                </button>
              ))
            )}
          </div>
          {/* 切换到文本模式的按钮 */}
          <button
            onClick={() => setMode('text')}
            className="mt-2 w-full rounded-2xl bg-cyan-300 py-2 text-sm font-black text-slate-950 transition active:scale-[0.99]"
          >
            返回文本输入
          </button>
        </div>
      )}
    </div>
  );
}
