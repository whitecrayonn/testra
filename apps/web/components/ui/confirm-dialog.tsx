"use client";

import { Trash2, type LucideIcon } from "lucide-react";
import { Dialog } from "@/components/ui/dialog";

interface ConfirmDialogProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  description: string;
  confirmLabel?: string;
  cancelLabel?: string;
  icon?: LucideIcon;
  confirming?: boolean;
}

export function ConfirmDialog({
  open,
  onClose,
  onConfirm,
  title,
  description,
  confirmLabel = "Delete",
  cancelLabel = "Keep it",
  icon: Icon = Trash2,
  confirming = false,
}: ConfirmDialogProps) {
  return (
    <Dialog open={open} onClose={onClose} ariaLabel={title} maxWidth="400px" tone="danger">
      <div className="p-6">
        <div className="flex h-[42px] w-[42px] items-center justify-center rounded-2xl bg-fail-soft text-fail">
          <Icon className="h-[18px] w-[18px]" aria-hidden="true" />
        </div>
        <h2 className="m-0 mb-1.5 mt-[15px] text-[17px] font-semibold tracking-tight text-fg">{title}</h2>
        <p className="m-0 text-[13px] text-fg2">{description}</p>
        <div className="mt-[22px] flex justify-end gap-2.5">
          <button
            onClick={onClose}
            className="h-[38px] rounded-[13px] border border-hair-hi bg-transparent px-4 text-[13px] font-semibold text-fg2 transition-colors hover:bg-panel-hi"
          >
            {cancelLabel}
          </button>
          <button
            onClick={onConfirm}
            disabled={confirming}
            className="h-[38px] rounded-[13px] border-0 bg-fail px-[18px] text-[13px] font-bold text-white shadow-[0_10px_26px_-12px_rgba(255,92,122,0.7)] disabled:opacity-60"
          >
            {confirming ? "Deleting…" : confirmLabel}
          </button>
        </div>
      </div>
    </Dialog>
  );
}
