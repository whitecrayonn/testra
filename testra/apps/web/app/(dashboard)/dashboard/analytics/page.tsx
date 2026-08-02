"use client";

import { useEffect, useState } from "react";
import { DashboardAnalytics } from "@/features/analytics/components/DashboardAnalytics";

interface AnalyticsState {
  workspaceId: string | null;
  projectId: string | null;
}

export default function AnalyticsPage() {
  const [state, setState] = useState<AnalyticsState | null>(null);

  useEffect(() => {
    setState({
      workspaceId: localStorage.getItem("testra_workspace_id"),
      projectId: localStorage.getItem("testra_project_id"),
    });
  }, []);

  if (!state) {
    return <div className="h-40 animate-pulse rounded-[20px] bg-panel" />;
  }

  if (!state.workspaceId) {
    return (
      <div className="animate-pop rounded-[20px] border border-dashed border-hair-hi bg-panel p-10 text-center shadow-glass">
        <p className="text-[13px] text-fg2">Select a workspace to see analytics.</p>
      </div>
    );
  }

  return (
    <DashboardAnalytics
      workspaceId={state.workspaceId}
      projectId={state.projectId || undefined}
      title="Analytics"
      description="Test execution, defects, and team activity across your workspace."
    />
  );
}
