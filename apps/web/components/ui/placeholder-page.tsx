import { LucideIcon, Check, X } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";

interface PlannedFeature {
  label: string;
  complete?: boolean;
}

interface PlaceholderPageProps {
  title: string;
  description: string;
  icon?: LucideIcon;
  status: string;
  releasePhase: string;
  features: PlannedFeature[];
  primaryCta?: { label: string; href?: string; onClick?: () => void };
  secondaryCta?: { label: string; href?: string; onClick?: () => void };
  breadcrumbs?: { label: string; href?: string }[];
}

function Cta({
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
    <Button variant={variant} onClick={onClick} disabled={!onClick}>
      {label}
    </Button>
  );
}

export function PlaceholderPage({
  title,
  description,
  icon: Icon,
  status,
  releasePhase,
  features,
  primaryCta,
  secondaryCta,
  breadcrumbs,
}: PlaceholderPageProps) {
  return (
    <div className="space-y-6">
      <PageHeader
        title={title}
        description={description}
        breadcrumbs={breadcrumbs}
        actions={primaryCta && <Cta label={primaryCta.label} href={primaryCta.href} onClick={primaryCta.onClick} variant="primary" />}
      />

      <Card className="overflow-hidden">
        <CardHeader className="border-b border-slate-200">
          <div className="flex items-center gap-3">
            {Icon && (
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-brand-50 text-brand-600">
                <Icon className="h-6 w-6" aria-hidden="true" />
              </div>
            )}
            <div>
              <CardTitle>{title}</CardTitle>
              <p className="text-sm text-slate-500">{description}</p>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-6 p-6">
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <p className="text-xs font-medium uppercase tracking-wide text-slate-500">Status</p>
              <p className="mt-1 text-sm font-medium text-slate-900">{status}</p>
            </div>
            <div>
              <p className="text-xs font-medium uppercase tracking-wide text-slate-500">Planned release</p>
              <p className="mt-1 text-sm font-medium text-slate-900">{releasePhase}</p>
            </div>
          </div>

          <div>
            <p className="text-xs font-medium uppercase tracking-wide text-slate-500">Features</p>
            <ul className="mt-3 space-y-2">
              {features.map((feature) => (
                <li
                  key={feature.label}
                  className="flex items-start gap-2 text-sm text-slate-700"
                >
                  <span className="mt-0.5">
                    {feature.complete ? (
                      <Check className="h-4 w-4 text-green-600" aria-hidden="true" />
                    ) : (
                      <X className="h-4 w-4 text-slate-300" aria-hidden="true" />
                    )}
                  </span>
                  {feature.label}
                </li>
              ))}
            </ul>
          </div>

          {primaryCta && !primaryCta.href && !primaryCta.onClick && (
            <p className="text-sm text-slate-600">
              {primaryCta.label} will be available once this module ships.
            </p>
          )}
        </CardContent>
      </Card>

      {secondaryCta && (
        <Cta label={secondaryCta.label} href={secondaryCta.href} onClick={secondaryCta.onClick} variant="secondary" />
      )}
    </div>
  );
}
