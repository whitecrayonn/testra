"use client";

import { type ReactNode } from "react";
import { createPortal } from "react-dom";
import { X, type LucideIcon } from "lucide-react";
import { usePortalOverlay } from "@/lib/hooks/use-portal-overlay";

interface DialogProps {
  open: boolean;
  onClose: () => void;
  children: ReactNode;
  maxWidth?: string;
  ariaLabel: string;
  tone?: "default" | "danger";
}

export function Dialog({ open, onClose, children, maxWidth = "520px", ariaLabel, tone = "default" }: DialogProps) {
  const panelRef = usePortalOverlay(open, onClose);

  if (!open) return null;

  return createPortal(
    <div
      className="fixed inset-0 z-[72] flex animate-fade-in items-center justify-center bg-[rgba(5,6,14,0.58)] p-6 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-label={ariaLabel}
        tabIndex={-1}
        onClick={(e) => e.stopPropagation()}
        style={{
          maxWidth,
          boxShadow: tone === "danger" ? "0 30px 70px -26px rgba(255,60,100,.35)" : "var(--shadow)",
        }}
        className={`w-full animate-pop overflow-hidden rounded-[24px] border bg-glass backdrop-blur-[30px] backdrop-saturate-150 outline-none ${
          tone === "danger" ? "border-[rgba(255,92,122,0.3)]" : "border-hair-hi"
        }`}
      >
        {children}
      </div>
    </div>,
    document.body,
  );
}

export function DialogHeader({
  icon: Icon,
  title,
  subtitle,
  onClose,
}: {
  icon?: LucideIcon;
  title: string;
  subtitle?: string;
  onClose: () => void;
}) {
  return (
    <div className="flex items-center gap-3 px-[22px] pt-5">
      {Icon && (
        <div className="flex h-[34px] w-[34px] flex-none items-center justify-center rounded-xl bg-acc-soft text-acc">
          <Icon className="h-4 w-4" aria-hidden="true" />
        </div>
      )}
      <div className="min-w-0 flex-1">
        <h2 className="m-0 text-[17px] font-semibold tracking-tight text-fg">{title}</h2>
        {subtitle && <p className="mt-0.5 text-xs text-fg3">{subtitle}</p>}
      </div>
      <button
        onClick={onClose}
        className="flex h-[30px] w-[30px] flex-none items-center justify-center rounded-[10px] text-fg3 transition-colors hover:bg-panel-hi"
        aria-label="Close"
      >
        <X className="h-[15px] w-[15px]" />
      </button>
    </div>
  );
}

export function DialogBody({ children }: { children: ReactNode }) {
  return <div className="flex flex-col gap-3.5 px-[22px] py-5">{children}</div>;
}

export function DialogFooter({ children }: { children: ReactNode }) {
  return <div className="flex justify-end gap-2.5 px-[22px] pb-5">{children}</div>;
}
