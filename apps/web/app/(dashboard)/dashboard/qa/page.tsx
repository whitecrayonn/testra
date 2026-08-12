"use client";

import { DashboardShell } from "@/components/dashboard/dashboard-shell";

export default function QADashboardPage() {
  return (
    <DashboardShell
      title="QA Dashboard"
      description="Manual QA execution activity, top failures, and team productivity."
      source="manual"
    />
  );
}
