'use client';

import { ReactNode, useRef, useState } from 'react';

interface PullToRefreshProps {
  children: ReactNode;
  onRefresh: () => Promise<void> | void;
  className?: string;
}

const TRIGGER_DISTANCE = 76;
const MAX_PULL_DISTANCE = 104;

export function PullToRefresh({ children, onRefresh, className = '' }: PullToRefreshProps) {
  const startYRef = useRef<number | null>(null);
  const pullingRef = useRef(false);
  const pullDistanceRef = useRef(0);
  const [pullDistance, setPullDistance] = useState(0);
  const [refreshing, setRefreshing] = useState(false);

  const reset = () => {
    startYRef.current = null;
    pullingRef.current = false;
    pullDistanceRef.current = 0;
    setPullDistance(0);
  };

  const handleTouchStart = (event: React.TouchEvent<HTMLDivElement>) => {
    if (refreshing || window.scrollY > 0) {
      return;
    }

    startYRef.current = event.touches[0]?.clientY ?? null;
    pullingRef.current = true;
  };

  const handleTouchMove = (event: React.TouchEvent<HTMLDivElement>) => {
    if (!pullingRef.current || startYRef.current === null || refreshing) {
      return;
    }

    const currentY = event.touches[0]?.clientY ?? startYRef.current;
    const delta = currentY - startYRef.current;
    if (delta <= 0) {
      pullDistanceRef.current = 0;
      setPullDistance(0);
      return;
    }

    const nextDistance = Math.min(delta * 0.55, MAX_PULL_DISTANCE);
    pullDistanceRef.current = nextDistance;
    setPullDistance(nextDistance);
  };

  const handleTouchEnd = async () => {
    if (!pullingRef.current) {
      return;
    }

    const shouldRefresh = pullDistanceRef.current >= TRIGGER_DISTANCE;
    reset();

    if (!shouldRefresh || refreshing) {
      return;
    }

    try {
      setRefreshing(true);
      await onRefresh();
    } finally {
      setRefreshing(false);
    }
  };

  const statusText = refreshing
    ? '刷新中'
    : pullDistance >= TRIGGER_DISTANCE
      ? '松开刷新'
      : '下拉刷新';

  const indicatorVisible = refreshing || pullDistance > 8;
  const indicatorOffset = refreshing ? 54 : Math.max(0, pullDistance - 20);

  return (
    <div
      data-pull-to-refresh="true"
      className={`relative overscroll-y-contain ${className}`}
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={() => void handleTouchEnd()}
      onTouchCancel={reset}
    >
      <div
        className={`pointer-events-none fixed left-1/2 top-3 z-50 flex -translate-x-1/2 items-center gap-2 rounded-full border border-cyan-300/20 bg-slate-950/95 px-3 py-2 text-xs font-semibold text-cyan-100 shadow-[0_0_28px_rgba(34,211,238,0.18)] backdrop-blur transition ${
          indicatorVisible ? 'opacity-100' : 'opacity-0'
        }`}
        style={{ transform: `translate(-50%, ${indicatorOffset}px)` }}
      >
        <span className={`h-2 w-2 rounded-full bg-cyan-300 ${refreshing ? 'animate-pulse' : ''}`} />
        {statusText}
      </div>
      <div
        style={{
          transform: pullDistance > 0 && !refreshing ? `translateY(${Math.min(pullDistance, 36)}px)` : undefined,
          transition: pullDistance === 0 ? 'transform 160ms ease-out' : undefined,
        }}
      >
        {children}
      </div>
    </div>
  );
}
