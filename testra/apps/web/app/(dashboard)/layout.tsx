"use client";

import { Sidebar } from "@/components/dashboard/sidebar";
import { Header } from "@/components/dashboard/header";
import { AmbientBackground } from "@/components/dashboard/ambient-background";
import { RouteGuard } from "@/components/auth/route-guard";
import { WorkspaceProvider } from "@/components/providers/workspace-provider";
import { CommandPalette } from "@/components/command-palette";
import { GlobalSearch } from "@/components/global-search";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <RouteGuard requireAuth redirectTo="/login">
      <WorkspaceProvider>
        <AmbientBackground />
        <div className="relative z-[1] box-border flex h-screen gap-3.5 p-3.5">
          <Sidebar />
          <div className="flex min-w-0 flex-1 flex-col gap-3.5">
            <Header />
            <main className="min-h-0 flex-1 overflow-auto px-1 pb-6 pt-0.5">{children}</main>
          </div>
        </div>
        <CommandPalette />
        <GlobalSearch />
      </WorkspaceProvider>
    </RouteGuard>
  );
}
