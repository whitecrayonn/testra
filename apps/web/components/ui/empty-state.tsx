import { LucideIcon } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";

interface EmptyStateProps {
  icon?: LucideIcon;
  title: string;
  description: string;
  action?: { label: string; href?: string; onClick?: () => void };
  secondaryAction?: { label: string; href?: string; onClick?: () => void };
}

function ActionButton({
  label,
  href,
  onClick,
  variant,
}: {
  label: string;
  href?: string;
  onClick?: () => void;
  variant?: "primary" | "secondary";
}) {
  const className =
    "inline-flex items-center justify-center rounded-lg px-4 py-2 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50";
  if (href) {
    return variant === "secondary" ? (
      <Link
        href={href}
        className={`${className} bg-white text-slate-900 border border-slate-300 hover:bg-slate-50 focus-visible:ring-slate-400`}
      >
        {label}
      </Link>
    ) : (
      <Link
        href={href}
        className={`${className} bg-brand-600 text-white hover:bg-brand-700 focus-visible:ring-brand-500`}
      >
        {label}
      </Link>
    );
  }
  return (
    <Button variant={variant} onClick={onClick}>
      {label}
    </Button>
  );
}

export function EmptyState({ icon: Icon, title, description, action, secondaryAction }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center rounded-xl border border-slate-200 bg-white p-12 text-center shadow-sm">
      {Icon && <Icon className="mb-4 h-12 w-12 text-slate-300" aria-hidden="true" />}
      <h3 className="text-lg font-semibold text-slate-900">{title}</h3>
      <p className="mt-1 max-w-sm text-sm text-slate-500">{description}</p>
      {(action || secondaryAction) && (
        <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
          {action && (
            <ActionButton
              label={action.label}
              href={action.href}
              onClick={action.onClick}
              variant="primary"
            />
          )}
          {secondaryAction && (
            <ActionButton
              label={secondaryAction.label}
              href={secondaryAction.href}
              onClick={secondaryAction.onClick}
              variant="secondary"
            />
          )}
        </div>
      )}
    </div>
  );
}
