"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { PlayCircle, TestTube, FolderKanban, Settings, Zap, ChevronRight, Activity } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { LinkButton } from "@/components/ui/link-button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { getSummary, getTrends } from "@/features/analytics/api";
import type { Summary, TrendPoint } from "@/types/analytics";

interface DashboardState {
  workspaceId: string | null;
  workspaceName: string | null;
  projectId: string | null;
  projectName: string | null;
}

export default function DashboardPage() {
  const [state, setState] = useState<DashboardState | null>(null);
  const [mounted, setMounted] = useState(false);
  const [summary, setSummary] = useState<Summary | null>(null);
  const [trends, setTrends] = useState<TrendPoint[]>([]);
  const [loadingAnalytics, setLoadingAnalytics] = useState(true);
  const [analyticsError, setAnalyticsError] = useState<string | null>(null);

  useEffect(() => {
    setMounted(true);
    if (typeof window !== "undefined") {
      const workspaceId = localStorage.getItem("testra_workspace_id");
      const projectId = localStorage.getItem("testra_project_id");
      setState({
        workspaceId,
        workspaceName: localStorage.getItem("testra_workspace_name"),
        projectId,
        projectName: localStorage.getItem("testra_project_name"),
      });

      if (workspaceId) {
        Promise.all([getSummary(workspaceId, projectId || undefined), getTrends(workspaceId, projectId || undefined)])
          .then(([s, t]) => {
            setSummary(s);
            setTrends(t.slice(-7));
          })
          .catch((err) => setAnalyticsError(err instanceof Error ? err.message : "Failed to load analytics"))
          .finally(() => setLoadingAnalytics(false));
      } else {
        setLoadingAnalytics(false);
      }
    }
  }, []);

  if (!mounted) {
    return (
      <div className="space-y-6">
        <PageHeader title="Dashboard" description="Loading workspace..." />
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Card key={i} className="h-28 animate-pulse bg-slate-100">
              <div aria-hidden="true" />
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Dashboard"
        description="Overview of your workspace, projects, and recent activity."
      />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Link href="/dashboard/test-runs" className="block">
          <Card className="h-full transition-colors hover:border-brand-300 hover:bg-brand-50">
            <CardContent className="flex h-full flex-col justify-between p-5">
              <div className="flex items-center justify-between">
                <PlayCircle className="h-8 w-8 text-brand-600" aria-hidden="true" />
                <ChevronRight className="h-4 w-4 text-slate-400" aria-hidden="true" />
              </div>
              <div>
                <p className="text-sm font-medium text-slate-600">Test Runs</p>
                <p className="text-2xl font-bold text-slate-900">View runs</p>
              </div>
            </CardContent>
          </Card>
        </Link>

        <Link href="/dashboard/test-cases" className="block">
          <Card className="h-full transition-colors hover:border-brand-300 hover:bg-brand-50">
            <CardContent className="flex h-full flex-col justify-between p-5">
              <div className="flex items-center justify-between">
                <TestTube className="h-8 w-8 text-brand-600" aria-hidden="true" />
                <ChevronRight className="h-4 w-4 text-slate-400" aria-hidden="true" />
              </div>
              <div>
                <p className="text-sm font-medium text-slate-600">Test Cases</p>
                <p className="text-2xl font-bold text-slate-900">Manage cases</p>
              </div>
            </CardContent>
          </Card>
        </Link>

        <Link href="/dashboard/projects" className="block">
          <Card className="h-full transition-colors hover:border-brand-300 hover:bg-brand-50">
            <CardContent className="flex h-full flex-col justify-between p-5">
              <div className="flex items-center justify-between">
                <FolderKanban className="h-8 w-8 text-brand-600" aria-hidden="true" />
                <ChevronRight className="h-4 w-4 text-slate-400" aria-hidden="true" />
              </div>
              <div>
                <p className="text-sm font-medium text-slate-600">Projects</p>
                <p className="text-2xl font-bold text-slate-900">Select project</p>
              </div>
            </CardContent>
          </Card>
        </Link>

        <Link href="/dashboard/settings" className="block">
          <Card className="h-full transition-colors hover:border-brand-300 hover:bg-brand-50">
            <CardContent className="flex h-full flex-col justify-between p-5">
              <div className="flex items-center justify-between">
                <Settings className="h-8 w-8 text-brand-600" aria-hidden="true" />
                <ChevronRight className="h-4 w-4 text-slate-400" aria-hidden="true" />
              </div>
              <div>
                <p className="text-sm font-medium text-slate-600">Settings</p>
                <p className="text-2xl font-bold text-slate-900">Configure</p>
              </div>
            </CardContent>
          </Card>
        </Link>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5 text-brand-600" />
            Analytics summary
          </CardTitle>
        </CardHeader>
        <CardContent>
          {analyticsError ? (
            <p className="text-sm text-red-600">{analyticsError}</p>
          ) : loadingAnalytics ? (
            <div className="grid gap-4 sm:grid-cols-5">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="h-16 animate-pulse rounded-lg bg-slate-100" />
              ))}
            </div>
          ) : summary ? (
            <div className="grid gap-4 sm:grid-cols-5">
              <Metric label="Total runs" value={summary.total_runs} />
              <Metric label="Passed" value={summary.passed} variant="success" />
              <Metric label="Failed" value={summary.failed} variant="danger" />
              <Metric label="Skipped" value={summary.skipped} />
              <Metric label="Blocked" value={summary.blocked} />
            </div>
          ) : (
            <p className="text-sm text-slate-500">Select a workspace to see analytics.</p>
          )}

          {trends.length > 0 && (
            <div className="mt-6">
              <h3 className="mb-3 text-sm font-medium text-slate-900">Recent trends</h3>
              <div className="space-y-2">
                {trends.map((point) => (
                  <div key={point.date} className="flex items-center justify-between rounded-lg border border-slate-100 p-3 text-sm">
                    <span className="text-slate-500">{point.date}</span>
                    <span className="font-medium">{point.total_runs} runs</span>
                    <span className="text-green-600">{point.passed} passed</span>
                    <span className="text-red-600">{point.failed} failed</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Quick actions</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-3">
              <LinkButton href="/dashboard/test-cases/new">
                <Zap className="mr-2 h-4 w-4" aria-hidden="true" />
                New test case
              </LinkButton>
              <LinkButton href="/dashboard/test-runs/new" variant="secondary">
                <PlayCircle className="mr-2 h-4 w-4" aria-hidden="true" />
                New test run
              </LinkButton>
              <LinkButton href="/dashboard/projects" variant="secondary">
                Manage projects
              </LinkButton>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Workspace context</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-slate-500">Workspace</span>
              <span className="font-medium text-slate-900">
                {state?.workspaceName || state?.workspaceId || "Not selected"}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-slate-500">Project</span>
              {state?.projectId ? (
                <Badge variant="success">{state.projectName || state.projectId.slice(0, 8)}</Badge>
              ) : (
                <Badge variant="warning">No project</Badge>
              )}
            </div>
            {!state?.projectId && (
              <p className="text-xs text-slate-500">
                Select a project in <Link href="/dashboard/projects" className="text-brand-600 hover:underline">Projects</Link> to enable test case and run creation.
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      {!state?.projectId && (
        <EmptyState
          icon={FolderKanban}
          title="Get started by selecting a project"
          description="Your workspace is ready. Choose or create a project, then add test cases and runs."
          action={{ label: "Go to Projects", href: "/dashboard/projects" }}
          secondaryAction={{ label: "Create test case", href: "/dashboard/test-cases/new" }}
        />
      )}
    </div>
  );
}

function Metric({
  label,
  value,
  variant,
}: {
  label: string;
  value: number;
  variant?: "success" | "danger";
}) {
  const color =
    variant === "success" ? "text-green-600" : variant === "danger" ? "text-red-600" : "text-slate-900";
  return (
    <div className="rounded-lg border border-slate-200 p-4">
      <p className="text-sm text-slate-500">{label}</p>
      <p className={`text-2xl font-bold ${color}`}>{value}</p>
    </div>
  );
}
