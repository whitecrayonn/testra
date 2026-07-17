"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

const settingsItems = [
  { href: "/dashboard/settings", label: "General" },
  { href: "/dashboard/settings/profile", label: "Profile" },
  { href: "/dashboard/settings/security", label: "Security" },
  { href: "/dashboard/settings/organization", label: "Organization" },
  { href: "/dashboard/settings/workspace", label: "Workspace" },
  { href: "/dashboard/settings/members", label: "Members" },
  { href: "/dashboard/settings/roles", label: "Roles" },
  { href: "/dashboard/settings/api-keys", label: "API Keys" },
  { href: "/dashboard/settings/notifications", label: "Notifications" },
  { href: "/dashboard/settings/audit-logs", label: "Audit Logs" },
  { href: "/dashboard/settings/billing", label: "Billing" },
];

export function SettingsNav() {
  const pathname = usePathname();

  return (
    <nav
      aria-label="Settings navigation"
      className="flex flex-wrap gap-2 border-b border-slate-200 pb-4"
    >
      {settingsItems.map((item) => {
        const active = pathname === item.href;
        return (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              "rounded-lg px-3 py-2 text-sm font-medium transition-colors",
              active
                ? "bg-brand-50 text-brand-700"
                : "text-slate-600 hover:bg-slate-100 hover:text-slate-900",
            )}
            aria-current={active ? "page" : undefined}
          >
            {item.label}
          </Link>
        );
      })}
    </nav>
  );
}
