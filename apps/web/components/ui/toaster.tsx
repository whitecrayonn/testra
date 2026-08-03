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
      className="fixed bottom-[18px] right-[18px] z-[88] flex flex-col items-end gap-2.5"
    >
      {toasts.map((t: ToastType) => (
        <Toast key={t.id} id={t.id} message={t.message} type={t.type} onDismiss={dismiss} />
      ))}
    </div>
  );
}
