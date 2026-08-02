"use client";

import { usePathname } from "next/navigation";
import { Search, Compass, RotateCcw, Bell, ChevronRight } from "lucide-react";
import { useWorkspace } from "@/lib/hooks/use-workspace";
import { useUnreadCount } from "@/features/notifications/hooks/use-unread-count";

const SEGMENT_TITLES: Record<string, string> = {
  dashboard: "Dashboard",
  projects: "Projects",
  "test-cases": "Test Cases",
  "test-runs": "Runs",
  defects: "Defects",
  "api-tests": "API Tests",
  automation: "Automation",
  "flaky-tests": "Flaky tests",
  notifications: "Notifications",
  settings: "Settings",
  analytics: "Analytics",
};

function screenTitleFromPath(pathname: string): string {
  const segments = pathname.split("/").filter(Boolean);
  for (let i = segments.length - 1; i >= 0; i--) {
    const title = SEGMENT_TITLES[segments[i]];
    if (title) return title;
  }
  return "Dashboard";
}

export function Header() {
  const pathname = usePathname();
  const { workspace } = useWorkspace();
  const unreadCount = useUnreadCount();
  const screenTitle = screenTitleFromPath(pathname);

  function openPalette() {
    document.dispatchEvent(new CustomEvent("open-command-palette"));
  }

  return (
    <header className="flex animate-rise items-center gap-3.5 rounded-[20px] border border-hair bg-glass p-2.5 px-3.5 shadow-glass backdrop-blur-[26px] backdrop-saturate-150">
      <div className="flex items-center gap-1.5 text-[12.5px] text-fg3">
        <span>{workspace?.name ?? "Testra"}</span>
        <ChevronRight className="h-3.5 w-3.5" aria-hidden="true" />
        <span className="font-semibold text-fg">{screenTitle}</span>
      </div>

      <button
        onClick={openPalette}
        className="flex h-[34px] max-w-[420px] flex-1 items-center gap-2 rounded-xl border border-hair bg-panel px-3 text-[12.5px] text-fg3 transition-colors hover:border-hair-hi hover:bg-panel-hi"
      >
        <Search className="h-3.5 w-3.5" aria-hidden="true" />
        <span className="flex-1 truncate text-left">Search or run a command…</span>
        <span className="rounded-md border border-hair-hi bg-panel px-1.5 py-0.5 font-mono text-[10px]">⌘K</span>
      </button>

      <div className="flex-1" />

      <button
        className="flex h-8 items-center gap-1.5 rounded-xl border border-hair bg-panel px-3 text-xs font-medium text-fg2 transition-colors hover:bg-panel-hi hover:text-fg"
        title="Tour"
      >
        <Compass className="h-3.5 w-3.5" aria-hidden="true" />
        Tour
      </button>
      <button
        className="flex h-8 w-8 items-center justify-center rounded-xl border border-hair bg-panel text-fg2 transition-colors hover:bg-panel-hi"
        title="Replay loading"
      >
        <RotateCcw className="h-3.5 w-3.5" aria-hidden="true" />
      </button>
      <button
        className="relative flex h-8 w-8 items-center justify-center rounded-xl border border-hair bg-panel text-fg2 transition-colors hover:bg-panel-hi"
        title="Notifications"
      >
        <Bell className="h-3.5 w-3.5" aria-hidden="true" />
        {unreadCount > 0 && (
          <span className="absolute right-1.5 top-1.5 h-1.5 w-1.5 animate-dot-pulse rounded-full bg-fail" />
        )}
      </button>
    </header>
  );
}
