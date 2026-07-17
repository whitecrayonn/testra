"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { PlayCircle, TestTube, FolderKanban, Settings, Zap, ChevronRight } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { LinkButton } from "@/components/ui/link-button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";

interface DashboardState {
  workspaceId: string | null;
  workspaceName: string | null;
  projectId: string | null;
  projectName: string | null;
}

export default function DashboardPage() {
  const [state, setState] = useState<DashboardState | null>(null);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    if (typeof window !== "undefined") {
      setState({
        workspaceId: localStorage.getItem("testra_workspace_id"),
        workspaceName: localStorage.getItem("testra_workspace_name"),
        projectId: localStorage.getItem("testra_project_id"),
        projectName: localStorage.getItem("testra_project_name"),
      });
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
