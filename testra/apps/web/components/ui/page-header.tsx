interface PageHeaderProps {
  title: string;
  description?: string;
  actions?: React.ReactNode;
  breadcrumbs?: { label: string; href?: string }[];
}

export function PageHeader({ title, description, actions, breadcrumbs }: PageHeaderProps) {
  return (
    <div className="space-y-4">
      {breadcrumbs && breadcrumbs.length > 0 && (
        <nav aria-label="Breadcrumb" className="flex items-center gap-2 text-sm text-slate-500">
          {breadcrumbs.map((crumb, index) => (
            <span key={crumb.label + index} className="flex items-center gap-2">
              {index > 0 && <span aria-hidden="true">/</span>}
              {crumb.href ? (
                <a href={crumb.href} className="hover:text-brand-600">
                  {crumb.label}
                </a>
              ) : (
                <span className="text-slate-900">{crumb.label}</span>
              )}
            </span>
          ))}
        </nav>
      )}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">{title}</h1>
          {description && <p className="mt-1 text-sm text-slate-500">{description}</p>}
        </div>
        {actions && <div className="flex items-center gap-2">{actions}</div>}
      </div>
    </div>
  );
}
