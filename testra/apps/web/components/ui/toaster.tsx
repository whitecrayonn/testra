"use client";

import { useToast } from "@/lib/hooks/use-toast";
import { Toast } from "@/components/ui/toast";
import type { Toast as ToastType } from "@/components/providers/toast-provider";

export function Toaster() {
  const { toasts, dismiss } = useToast();
  if (toasts.length === 0) return null;
  return (
    <div
      aria-live="polite"
      aria-atomic="false"
      className="fixed right-4 top-4 z-50 flex w-full max-w-sm flex-col gap-2"
    >
      {toasts.map((t: ToastType) => (
        <Toast key={t.id} id={t.id} message={t.message} type={t.type} onDismiss={dismiss} />
      ))}
    </div>
  );
}
