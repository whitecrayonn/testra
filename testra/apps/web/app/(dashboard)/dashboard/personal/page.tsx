"use client";

import { DashboardShell } from "@/components/dashboard/dashboard-shell";

export default function PersonalDashboardPage() {
  return (
    <DashboardShell
      title="Personal Dashboard"
      description="Your personal execution activity and contributions."
      personal
    />
  );
}
