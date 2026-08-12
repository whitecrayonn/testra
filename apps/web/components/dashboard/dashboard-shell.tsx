"use client";

import { useEffect, useState } from "react";
import { Loader2 } from "lucide-react";
import { DashboardAnalytics } from "@/features/analytics/components/DashboardAnalytics";
import { apiFetch } from "@/lib/api";

interface DashboardShellProps {
  title: string;
  description: string;
  source?: string;
  personal?: boolean;
}

interface UserResponse {
  id: string;
  email: string;
  name: string;
}

export function DashboardShell({ title, description, source, personal }: DashboardShellProps) {
  const [workspaceId, setWorkspaceId] = useState<string | null>(null);
  const [projectId, setProjectId] = useState<string | null>(null);
  const [tester, setTester] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const ws = localStorage.getItem("testra_workspace_id");
    const pj = localStorage.getItem("testra_project_id");
    setWorkspaceId(ws);
    setProjectId(pj);
    if (personal) {
      apiFetch<UserResponse>("/api/v1/auth/me")
        .then((user) => setTester(user.id))
        .catch(() => setTester(undefined))
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, [personal]);

  if (loading) {
    return (
      <div className="flex h-96 items-center justify-center text-slate-500">
        <Loader2 className="mr-2 h-5 w-5 animate-spin" aria-hidden="true" />
        Loading...
      </div>
    );
  }

  if (!workspaceId) {
    return (
      <div className="flex h-96 items-center justify-center text-sm text-slate-500">Select a workspace to view this dashboard.</div>
    );
  }

  return (
    <DashboardAnalytics
      workspaceId={workspaceId}
      projectId={projectId || undefined}
      title={title}
      description={description}
      source={source}
      tester={tester}
    />
  );
}
