"use client";

import { ToastProvider } from "@/components/providers/toast-provider";
import { NetworkStatusProvider } from "@/components/providers/network-status-provider";
import { Toaster } from "@/components/ui/toaster";
import { GlobalLoadingBar } from "@/components/ui/global-loading-bar";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ToastProvider>
      <NetworkStatusProvider>
        <GlobalLoadingBar />
        {children}
        <Toaster />
      </NetworkStatusProvider>
    </ToastProvider>
  );
}
