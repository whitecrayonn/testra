"use client";

import { Sidebar } from "@/components/dashboard/sidebar";
import { RouteGuard } from "@/components/auth/route-guard";
import { WorkspaceProvider } from "@/components/providers/workspace-provider";
import { CommandPalette } from "@/components/command-palette";
import { GlobalSearch } from "@/components/global-search";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <RouteGuard requireAuth redirectTo="/login">
      <WorkspaceProvider>
        <div className="flex h-screen">
          <Sidebar />
          <main className="flex-1 overflow-auto bg-slate-50 p-6 dark:bg-slate-950">{children}</main>
        </div>
        <CommandPalette />
        <GlobalSearch />
      </WorkspaceProvider>
    </RouteGuard>
  );
}
