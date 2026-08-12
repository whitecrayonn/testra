"use client";

import { RouteGuard } from "@/components/auth/route-guard";
import { CreateWorkspaceForm } from "@/components/workspace/create-workspace-form";

export default function OnboardingWorkspacePage() {
  return (
    <RouteGuard requireAuth redirectTo="/login">
      <div className="flex min-h-screen flex-col items-center justify-center bg-slate-50 px-4 dark:bg-slate-950">
        <CreateWorkspaceForm />
      </div>
    </RouteGuard>
  );
}
