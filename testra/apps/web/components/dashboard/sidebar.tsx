"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { getUnreadCount } from "@/features/notifications/api";
import {
  LayoutDashboard,
  FolderKanban,
  TestTube,
  PlayCircle,
  Bug,
  Bell,
  Settings,
  LogOut,
} from "lucide-react";

const navItems = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/dashboard/projects", label: "Projects", icon: FolderKanban },
  { href: "/dashboard/test-cases", label: "Test Cases", icon: TestTube },
  { href: "/dashboard/test-runs", label: "Runs", icon: PlayCircle },
  { href: "/dashboard/defects", label: "Defects", icon: Bug },
  { href: "/dashboard/notifications", label: "Notifications", icon: Bell },
  { href: "/dashboard/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    async function poll() {
      try {
        const { unread_count } = await getUnreadCount();
        setUnreadCount(unread_count);
      } catch {
        setUnreadCount(0);
      }
    }
    poll();
    const id = setInterval(poll, 30000);
    return () => clearInterval(id);
  }, []);

  return (
    <aside className="flex h-screen w-60 flex-col border-r border-slate-200 bg-white">
      <div className="flex h-14 items-center border-b border-slate-200 px-4">
        <Link href="/dashboard" className="text-lg font-bold text-brand-600">
          Testra
        </Link>
      </div>
      <nav className="flex-1 space-y-1 px-2 py-4">
        {navItems.map((item) => {
          const Icon = item.icon;
          const active =
            pathname === item.href ||
            pathname.startsWith(item.href + "/") ||
            (item.href === "/dashboard/settings" && pathname.startsWith("/dashboard/settings"));
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                active
                  ? "bg-brand-50 text-brand-700"
                  : "text-slate-600 hover:bg-slate-100 hover:text-slate-900",
              )}
              aria-current={active ? "page" : undefined}
            >
              <Icon className="h-4 w-4" aria-hidden="true" />
              {item.label}
              {item.href === "/dashboard/notifications" && unreadCount > 0 && (
                <Badge variant="danger" className="ml-auto px-1.5 py-0.5 text-[10px]">
                  {unreadCount > 99 ? "99+" : unreadCount}
                </Badge>
              )}
            </Link>
          );
        })}
      </nav>
      <div className="border-t border-slate-200 p-2">
        <Link
          href="/login"
          className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-slate-600 hover:bg-slate-100"
        >
          <LogOut className="h-4 w-4" />
          Sign out
        </Link>
      </div>
    </aside>
  );
}
