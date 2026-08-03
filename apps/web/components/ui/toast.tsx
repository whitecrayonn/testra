import { X, CheckCircle2, XCircle, AlertTriangle, Info } from "lucide-react";
import type { ToastType } from "@/components/providers/toast-provider";

const TONE: Record<ToastType, { badgeBg: string; fg: string; barBg: string; icon: typeof CheckCircle2 }> = {
  success: { badgeBg: "bg-pass-soft", fg: "text-pass", barBg: "bg-pass", icon: CheckCircle2 },
  error: { badgeBg: "bg-fail-soft", fg: "text-fail", barBg: "bg-fail", icon: XCircle },
  warning: { badgeBg: "bg-warn-soft", fg: "text-warn", barBg: "bg-warn", icon: AlertTriangle },
  info: { badgeBg: "bg-acc-soft", fg: "text-info", barBg: "bg-info", icon: Info },
};

interface ToastProps {
  id: string;
  message: string;
  type: ToastType;
  onDismiss?: (id: string) => void;
}

export function Toast({ id, message, type, onDismiss }: ToastProps) {
  const tone = TONE[type];
  const Icon = tone.icon;
  return (
    <div
      role="alert"
      className="relative flex min-w-[280px] animate-toast-in items-center gap-3 overflow-hidden rounded-2xl border border-hair-hi bg-glass p-3.5 shadow-glass backdrop-blur-[26px] backdrop-saturate-150"
    >
      <span className={`flex h-[26px] w-[26px] flex-none items-center justify-center rounded-lg ${tone.badgeBg} ${tone.fg}`}>
        <Icon className="h-3.5 w-3.5" aria-hidden="true" />
      </span>
      <p className="flex-1 text-[12.5px] font-medium text-fg">{message}</p>
      {onDismiss && (
        <button
          onClick={() => onDismiss(id)}
          className="flex-none text-fg3 transition-colors hover:text-fg"
          aria-label="Dismiss"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      )}
      <span className={`absolute inset-x-0 bottom-0 h-0.5 origin-left animate-toast-bar ${tone.barBg}`} />
    </div>
  );
}
