'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import BindPage from './components/BindPage';
import Terminal from './components/Terminal';

export default function Home() {
  const [deviceId, setDeviceId] = useState<string | null>(null);
  const [isMobile, setIsMobile] = useState(false);
  const router = useRouter();

  useEffect(() => {
    const savedDeviceId = localStorage.getItem('device_id');
    if (savedDeviceId) setDeviceId(savedDeviceId);

    // 检测移动端
    setIsMobile(window.innerWidth < 768);
    const handleResize = () => setIsMobile(window.innerWidth < 768);
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const handleBind = (id: string) => {
    localStorage.setItem('device_id', id);
    setDeviceId(id);
    router.refresh();
  };

  // 简化版：无需登录，直接判断是否有设备绑定
  if (!deviceId) {
    return <BindPage onBind={handleBind} />;
  }

  return (
    <div className="relative">
      <button
        onClick={() => {
          localStorage.removeItem('device_id');
          setDeviceId(null);
          router.refresh();
        }}
        className={`fixed z-50 text-gray-400 hover:text-white transition-colors ${
          isMobile
            ? 'top-2 right-2 px-2 py-1 text-xs'
            : 'top-4 right-4 px-3 py-1 text-xs'
        }`}
      >
        {'切换绑定码'}
      </button>
      <Terminal deviceId={deviceId} />
    </div>
  );
}
