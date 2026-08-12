"use client";

import { useEffect, useState, type ReactNode } from "react";

export function NetworkStatusProvider({ children }: { children: ReactNode }) {
  const [online, setOnline] = useState(true);

  useEffect(() => {
    const on = () => setOnline(true);
    const off = () => setOnline(false);
    setOnline(navigator.onLine);
    window.addEventListener("online", on);
    window.addEventListener("offline", off);
    return () => {
      window.removeEventListener("online", on);
      window.removeEventListener("offline", off);
    };
  }, []);

  return (
    <>
      {!online && (
        <div
          role="alert"
          className="fixed inset-x-0 top-0 z-50 bg-yellow-50 px-4 py-2 text-center text-sm font-medium text-yellow-800"
        >
          You are offline. Some features may be unavailable until the connection is restored.
        </div>
      )}
      {children}
    </>
  );
}
