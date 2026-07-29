import { X } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ToastType } from "@/components/providers/toast-provider";

const styles: Record<ToastType, string> = {
  success: "bg-green-50 text-green-800 border-green-200",
  error: "bg-red-50 text-red-800 border-red-200",
  warning: "bg-yellow-50 text-yellow-800 border-yellow-200",
  info: "bg-blue-50 text-blue-800 border-blue-200",
};

const icons: Record<ToastType, string> = {
  success: "✓",
  error: "✕",
  warning: "!",
  info: "i",
};

interface ToastProps {
  id: string;
  message: string;
  type: ToastType;
  onDismiss?: (id: string) => void;
}

export function Toast({ id, message, type, onDismiss }: ToastProps) {
  return (
    <div
      role="alert"
      className={cn(
        "flex items-start gap-3 rounded-lg border px-4 py-3 shadow-sm",
        styles[type],
      )}
    >
      <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-current/10 text-xs font-bold">
        {icons[type]}
      </span>
      <p className="flex-1 text-sm font-medium">{message}</p>
      {onDismiss && (
        <button
          onClick={() => onDismiss(id)}
          className="text-current/70 hover:text-current"
          aria-label="Dismiss"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  );
}
