"use client";

import { useEffect, useState } from "react";
import { subscribeToPendingRequests } from "@/lib/api";

// Debounces show/hide so a burst of fast requests (most navigations, most
// actions) reads as one smooth bar instead of a flicker.
const SHOW_DELAY_MS = 150;
const MIN_VISIBLE_MS = 400;

export function GlobalLoadingBar() {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    let showTimer: ReturnType<typeof setTimeout> | null = null;
    let hideTimer: ReturnType<typeof setTimeout> | null = null;
    let shownAt = 0;

    const unsubscribe = subscribeToPendingRequests((count) => {
      if (count > 0) {
        if (hideTimer) {
          clearTimeout(hideTimer);
          hideTimer = null;
        }
        if (!showTimer && !shownAt) {
          showTimer = setTimeout(() => {
            shownAt = Date.now();
            showTimer = null;
            setVisible(true);
          }, SHOW_DELAY_MS);
        }
        return;
      }

      if (showTimer) {
        clearTimeout(showTimer);
        showTimer = null;
        return;
      }
      if (shownAt) {
        const remaining = Math.max(0, MIN_VISIBLE_MS - (Date.now() - shownAt));
        hideTimer = setTimeout(() => {
          hideTimer = null;
          shownAt = 0;
          setVisible(false);
        }, remaining);
      }
    });

    return () => {
      unsubscribe();
      if (showTimer) clearTimeout(showTimer);
      if (hideTimer) clearTimeout(hideTimer);
    };
  }, []);

  if (!visible) return null;

  return (
    <div
      role="status"
      aria-live="polite"
      aria-label="Loading"
      className="fixed inset-x-0 top-0 z-[100] h-[3px] overflow-hidden bg-hair"
    >
      <div className="h-full w-1/3 animate-sheen bg-gradient-to-r from-transparent via-acc to-acc2 shadow-[0_0_10px_var(--acc)]" />
    </div>
  );
}
