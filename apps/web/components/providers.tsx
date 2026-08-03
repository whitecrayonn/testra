"use client";

import { ToastProvider } from "@/components/providers/toast-provider";
import { NetworkStatusProvider } from "@/components/providers/network-status-provider";
import { Toaster } from "@/components/ui/toaster";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ToastProvider>
      <NetworkStatusProvider>
        {children}
        <Toaster />
      </NetworkStatusProvider>
    </ToastProvider>
  );
}
