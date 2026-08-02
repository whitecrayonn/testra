"use client";

import { type ReactNode } from "react";
import { createPortal } from "react-dom";
import { X, type LucideIcon } from "lucide-react";
import { usePortalOverlay } from "@/lib/hooks/use-portal-overlay";

interface SlideOverProps {
  open: boolean;
  onClose: () => void;
  children: ReactNode;
  ariaLabel: string;
  width?: string;
}

export function SlideOver({ open, onClose, children, ariaLabel, width = "480px" }: SlideOverProps) {
  const panelRef = usePortalOverlay(open, onClose);

  if (!open) return null;

  return createPortal(
    <div
      className="fixed inset-0 z-[68] animate-fade-in bg-[rgba(5,6,14,0.5)] backdrop-blur-[3px]"
      onClick={onClose}
    >
      <div
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-label={ariaLabel}
        tabIndex={-1}
        onClick={(e) => e.stopPropagation()}
        style={{ width }}
        className="absolute bottom-3.5 right-3.5 top-3.5 flex animate-drawer-in flex-col overflow-hidden rounded-[24px] border border-hair-hi bg-glass shadow-glass backdrop-blur-[30px] backdrop-saturate-150 outline-none"
      >
        {children}
      </div>
    </div>,
    document.body,
  );
}

export function SlideOverHeader({
  icon: Icon,
  kicker,
  title,
  meta,
  tint,
  color,
  onClose,
}: {
  icon: LucideIcon;
  kicker: string;
  title: string;
  meta?: string;
  tint?: string;
  color?: string;
  onClose: () => void;
}) {
  return (
    <div className="flex items-start gap-3.5 border-b border-hair px-[22px] py-4">
      <div
        className="flex h-10 w-10 flex-none items-center justify-center rounded-2xl"
        style={{ background: tint ?? "var(--acc-soft)", color: color ?? "var(--acc)" }}
      >
        <Icon className="h-[18px] w-[18px]" aria-hidden="true" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="font-mono text-[10.5px] tracking-wide text-fg3">{kicker}</div>
        <h2 className="m-0 mt-1 text-[18px] font-semibold tracking-tight text-fg">{title}</h2>
        {meta && <div className="mt-1 text-xs text-fg3">{meta}</div>}
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

export function SlideOverBody({ children }: { children: ReactNode }) {
  return <div className="flex-1 overflow-y-auto px-[22px] py-[18px]">{children}</div>;
}

export function SlideOverFooter({ children }: { children: ReactNode }) {
  return <div className="flex gap-2.5 border-t border-hair px-[22px] py-4">{children}</div>;
}
