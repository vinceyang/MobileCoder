'use client';

import { useEffect, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import Terminal from '../components/Terminal';

export default function TerminalPage() {
  const [deviceId, setDeviceId] = useState('');
  const searchParams = useSearchParams();
  const router = useRouter();

  useEffect(() => {
    // Get device_id from URL query params or localStorage
    const deviceIdFromUrl = searchParams.get('device_id');
    const deviceIdFromStorage = localStorage.getItem('device_id');
    const deviceId = deviceIdFromUrl || deviceIdFromStorage;

    if (!deviceId) {
      router.push('/login');
      return;
    }
    setDeviceId(deviceId);
  }, [searchParams, router]);

  if (!deviceId) {
    return null;
  }

  return (
    <div>
      <Terminal deviceId={deviceId} />
    </div>
  );
}
