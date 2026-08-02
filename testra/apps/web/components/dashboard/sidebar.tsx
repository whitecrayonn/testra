"use client";

import { useLayoutEffect, useRef, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { cn } from "@/lib/utils";
import { useWorkspace } from "@/lib/hooks/use-workspace";
import { useUnreadCount } from "@/features/notifications/hooks/use-unread-count";
import { logout } from "@/lib/api";
import {
  LayoutDashboard,
  Activity,
  FolderKanban,
  TestTube,
  PlayCircle,
  Bug,
  Bell,
  Brain,
  Settings,
  LogOut,
  Globe,
  Bot,
  Moon,
  Sun,
  ChevronsUpDown,
} from "lucide-react";
import { useTheme } from "@/components/providers/theme-provider";

type NavEntry =
  | { kind: "label"; key: string; label: string }
  | { kind: "item"; key: string; href: string; label: string; icon: typeof LayoutDashboard; badgeCount?: number };

export function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { workspace, user } = useWorkspace();
  const { theme, setTheme } = useTheme();
  const unreadCount = useUnreadCount();

  const nav: NavEntry[] = [
    { kind: "label", key: "l-workspace", label: "WORKSPACE" },
    { kind: "item", key: "dashboard", href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
    { kind: "item", key: "analytics", href: "/dashboard/executive", label: "Analytics", icon: Activity },
    { kind: "item", key: "projects", href: "/dashboard/projects", label: "Projects", icon: FolderKanban },
    { kind: "label", key: "l-quality", label: "QUALITY" },
    { kind: "item", key: "test-cases", href: "/dashboard/test-cases", label: "Test Cases", icon: TestTube },
    { kind: "item", key: "test-runs", href: "/dashboard/test-runs", label: "Runs", icon: PlayCircle },
    { kind: "item", key: "defects", href: "/dashboard/defects", label: "Defects", icon: Bug },
    { kind: "item", key: "api-tests", href: "/dashboard/api-tests", label: "API Tests", icon: Globe },
    { kind: "item", key: "automation", href: "/dashboard/automation", label: "Automation", icon: Bot },
    { kind: "item", key: "flaky-tests", href: "/dashboard/flaky-tests", label: "Flaky tests", icon: Brain },
    { kind: "label", key: "l-system", label: "SYSTEM" },
    {
      kind: "item",
      key: "notifications",
      href: "/dashboard/notifications",
      label: "Notifications",
      icon: Bell,
      badgeCount: unreadCount,
    },
    { kind: "item", key: "settings", href: "/dashboard/settings", label: "Settings", icon: Settings },
  ];

  function isActive(href: string) {
    return (
      pathname === href ||
      pathname.startsWith(href + "/") ||
      (href === "/dashboard/settings" && pathname.startsWith("/dashboard/settings"))
    );
  }

  const itemRefs = useRef<Map<string, HTMLAnchorElement>>(new Map());
  const navRef = useRef<HTMLElement>(null);
  const [indicator, setIndicator] = useState<{ top: number; height: number } | null>(null);

  useLayoutEffect(() => {
    const activeEntry = nav.find((n) => n.kind === "item" && isActive(n.href)) as
      | Extract<NavEntry, { kind: "item" }>
      | undefined;
    if (!activeEntry || !navRef.current) {
      setIndicator(null);
      return;
    }
    const el = itemRefs.current.get(activeEntry.key);
    if (!el) return;
    setIndicator({ top: el.offsetTop, height: el.offsetHeight });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pathname]);

  async function signOut() {
    try {
      await logout();
    } catch {
      // Ignore logout failures; redirect to login regardless.
    }
    router.replace("/login");
  }

  return (
    <aside className="flex w-[252px] flex-shrink-0 animate-rise flex-col rounded-[22px] border border-hair bg-glass p-3.5 pt-4.5 shadow-glass backdrop-blur-[26px] backdrop-saturate-150">
      <div className="mb-4 flex items-center gap-2.5 px-1.5">
        <div className="flex h-[34px] w-[34px] flex-none items-end justify-center gap-[3px] rounded-[11px] bg-gradient-to-br from-acc to-acc2 pb-2 shadow-[0_6px_18px_-6px_var(--ring)]">
          <span className="h-2 w-[3px] rounded-sm bg-white/95" />
          <span className="h-3.5 w-[3px] rounded-sm bg-white/95" />
          <span className="h-[11px] w-[3px] rounded-sm bg-white/95" />
        </div>
        <div className="flex flex-col gap-px">
          <span className="text-[15px] font-bold tracking-[0.14em] text-fg">TESTRA</span>
          <span className="font-mono text-[9px] tracking-[0.1em] text-fg3">EVERY TEST · v1.4</span>
        </div>
      </div>

      <button className="mb-4 flex w-full items-center gap-2.5 rounded-[13px] border border-hair bg-panel px-2.5 py-2 text-left transition-colors hover:border-hair-hi hover:bg-panel-hi">
        <div className="flex h-6 w-6 items-center justify-center rounded-lg bg-acc-soft text-[11px] font-bold text-acc">
          {(workspace?.name ?? "TC").slice(0, 2).toUpperCase()}
        </div>
        <div className="min-w-0 flex-1">
          <div className="truncate text-[12.5px] font-semibold leading-tight text-fg">
            {workspace?.name ?? "Select workspace"}
          </div>
          <div className="truncate text-[10.5px] leading-tight text-fg3">Workspace</div>
        </div>
        <ChevronsUpDown className="h-3.5 w-3.5 flex-none text-fg3" aria-hidden="true" />
      </button>

      <nav ref={navRef} className="relative min-h-0 flex-1 overflow-y-auto overflow-x-hidden">
        {indicator && (
          <>
            <div
              className="pointer-events-none absolute left-0 right-0 h-[38px] rounded-xl border border-hair-hi bg-acc-soft transition-transform duration-[420ms] ease-[cubic-bezier(.34,1.32,.36,1)]"
              style={{ transform: `translateY(${indicator.top}px)` }}
            />
            <div
              className="pointer-events-none absolute left-0 h-4 w-[3px] rounded-[3px] bg-acc shadow-[0_0_12px_var(--acc)] transition-transform duration-[420ms] ease-[cubic-bezier(.34,1.32,.36,1)]"
              style={{ transform: `translateY(${indicator.top + (indicator.height - 16) / 2}px)` }}
            />
          </>
        )}

        {nav.map((entry) => {
          if (entry.kind === "label") {
            return (
              <div
                key={entry.key}
                className="px-2.5 pb-1.5 pt-3.5 font-mono text-[9px] tracking-[0.16em] text-fg3"
              >
                {entry.label}
              </div>
            );
          }
          const Icon = entry.icon;
          const active = isActive(entry.href);
          return (
            <Link
              key={entry.key}
              href={entry.href}
              ref={(el) => {
                if (el) itemRefs.current.set(entry.key, el);
                else itemRefs.current.delete(entry.key);
              }}
              aria-current={active ? "page" : undefined}
              className={cn(
                "relative mb-1 flex h-[38px] items-center gap-2.5 rounded-xl px-2.5 text-[13px] font-medium transition-colors hover:bg-panel",
                active ? "text-fg" : "text-fg2",
              )}
            >
              <Icon className="h-[15px] w-[15px] flex-none" aria-hidden="true" />
              <span className="flex-1">{entry.label}</span>
              {!!entry.badgeCount && entry.badgeCount > 0 && (
                <span className="rounded-full bg-fail-soft px-1.5 py-0.5 font-mono text-[9.5px] font-bold text-fail">
                  {entry.badgeCount > 99 ? "99+" : entry.badgeCount}
                </span>
              )}
            </Link>
          );
        })}
      </nav>

      <div className="mt-3 flex gap-1 rounded-[13px] border border-hair bg-panel p-1">
        <button
          onClick={() => setTheme("dark")}
          className={cn(
            "flex h-[30px] flex-1 items-center justify-center gap-1.5 rounded-[10px] text-[11.5px] font-semibold transition-colors",
            theme === "dark" ? "bg-acc text-white" : "text-fg2",
          )}
        >
          <Moon className="h-3.5 w-3.5" aria-hidden="true" />
          Dark
        </button>
        <button
          onClick={() => setTheme("light")}
          className={cn(
            "flex h-[30px] flex-1 items-center justify-center gap-1.5 rounded-[10px] text-[11.5px] font-semibold transition-colors",
            theme === "light" ? "bg-acc text-white" : "text-fg2",
          )}
        >
          <Sun className="h-3.5 w-3.5" aria-hidden="true" />
          Light
        </button>
      </div>

      <div className="mt-2.5 flex items-center gap-2.5 border-t border-hair pt-2">
        <div className="flex h-7 w-7 items-center justify-center rounded-[10px] bg-gradient-to-br from-acc2 to-acc text-[11px] font-bold text-[#08111a]">
          {(user?.name ?? user?.email ?? "U").slice(0, 2).toUpperCase()}
        </div>
        <div className="min-w-0 flex-1">
          <div className="truncate text-xs font-semibold text-fg">{user?.name ?? "Account"}</div>
          <div className="truncate text-[10px] text-fg3">{user?.email ?? "Signed in"}</div>
        </div>
        <button
          onClick={signOut}
          title="Sign out"
          className="flex h-7 w-7 items-center justify-center rounded-[9px] text-fg3 transition-colors hover:bg-panel-hi hover:text-fail"
        >
          <LogOut className="h-3.5 w-3.5" aria-hidden="true" />
        </button>
      </div>
    </aside>
  );
}
