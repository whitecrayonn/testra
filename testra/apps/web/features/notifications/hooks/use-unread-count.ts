"use client";

import { useEffect, useState } from "react";
import { getUnreadCount } from "@/features/notifications/api";

export function useUnreadCount(pollIntervalMs = 30000): number {
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    let cancelled = false;

    async function poll() {
      try {
        const { unread_count } = await getUnreadCount();
        if (!cancelled) setUnreadCount(unread_count);
      } catch {
        if (!cancelled) setUnreadCount(0);
      }
    }

    poll();
    const id = setInterval(poll, pollIntervalMs);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, [pollIntervalMs]);

  return unreadCount;
}
