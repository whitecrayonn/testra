"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Play, Plus, FileText, FolderKanban } from "lucide-react";
import { ProjectOnboarding } from "@/components/project-onboarding";
import { DashboardAnalytics } from "@/features/analytics/components/DashboardAnalytics";
import { ReleaseReadinessCard } from "@/components/dashboard/release-readiness-card";
import { LiveExecutionCard } from "@/components/dashboard/live-execution-card";
import { getMetrics } from "@/features/analytics/api";
import { listTestRuns } from "@/features/results/api";
import type { Metrics } from "@/types/analytics";
import type { TestRun } from "@/types/results";

interface DashboardState {
  workspaceId: string | null;
  workspaceName: string | null;
  projectId: string | null;
  projectName: string | null;
}

export default function DashboardPage() {
  const router = useRouter();
  const [state, setState] = useState<DashboardState | null>(null);
  const [mounted, setMounted] = useState(false);
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [loadingHero, setLoadingHero] = useState(true);

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
    }
  }, []);

  useEffect(() => {
    if (!state?.workspaceId) return;
    let cancelled = false;
    setLoadingHero(true);
    Promise.all([
      getMetrics({ workspace_id: state.workspaceId, project_id: state.projectId || undefined }),
      state.projectId ? listTestRuns(state.projectId, { limit: 5 }) : Promise.resolve({ data: [], meta: {} }),
    ])
      .then(([m, r]) => {
        if (cancelled) return;
        setMetrics(m);
        setRuns(r.data);
      })
      .catch(() => {
        if (!cancelled) {
          setMetrics(null);
          setRuns([]);
        }
      })
      .finally(() => {
        if (!cancelled) setLoadingHero(false);
      });
    return () => {
      cancelled = true;
    };
  }, [state?.workspaceId, state?.projectId]);

  if (!mounted) {
    return (
      <div className="flex flex-col gap-3.5">
        <div className="h-10 w-64 animate-pulse rounded-xl bg-panel" />
        <div className="grid grid-cols-2 gap-3.5">
          <div className="h-40 animate-pulse rounded-[20px] bg-panel" />
          <div className="h-40 animate-pulse rounded-[20px] bg-panel" />
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3.5">
      <div className="flex animate-rise-sm items-end justify-between gap-4">
        <div>
          <h1 className="m-0 text-[27px] font-bold tracking-tight text-fg">Dashboard</h1>
          <p className="mt-1.5 text-[13px] text-fg2">
            {state?.projectName || state?.workspaceName || "Select a workspace"} · Quality overview
          </p>
        </div>
        <div className="flex gap-2.5">
          <Link
            href="/dashboard/test-runs/new"
            aria-label="New run — Test Runs"
            className="flex h-[38px] items-center gap-2 rounded-[13px] border border-hair-hi bg-panel px-4 text-[13px] font-semibold text-fg transition-colors hover:bg-panel-hi"
          >
            <Play className="h-3.5 w-3.5" aria-hidden="true" />
            New run
          </Link>
          <Link
            href="/dashboard/test-cases/new"
            aria-label="New test case — Test Cases"
            className="flex h-[38px] items-center gap-2 rounded-[13px] bg-gradient-to-br from-acc to-acc2 px-4 text-[13px] font-bold text-white shadow-[0_10px_26px_-12px_var(--ring)] transition-transform hover:-translate-y-px"
          >
            <Plus className="h-3.5 w-3.5" aria-hidden="true" />
            New test case
          </Link>
        </div>
      </div>

      {state?.workspaceId ? (
        <>
          <div data-tour="hero" className="grid grid-cols-1 gap-3.5 lg:grid-cols-[1.55fr_1fr]">
            <ReleaseReadinessCard metrics={metrics} loading={loadingHero} />
            <LiveExecutionCard runs={runs} loading={loadingHero} />
          </div>

          <DashboardAnalytics workspaceId={state.workspaceId} projectId={state.projectId || undefined} />

          <div className="grid gap-3.5 lg:grid-cols-3">
            <div className="animate-pop rounded-[20px] border border-hair bg-panel p-5 shadow-glass lg:col-span-2">
              <h3 className="m-0 mb-3 text-sm font-semibold text-fg">Quick actions</h3>
              <div className="flex flex-wrap gap-2.5">
                <QuickAction href="/dashboard/test-runs/new" icon={Play} primary>
                  New test run
                </QuickAction>
                <QuickAction href="/dashboard/test-plans/new" icon={FileText}>
                  New test plan
                </QuickAction>
                <QuickAction href="/dashboard/projects" icon={FolderKanban}>
                  Manage projects
                </QuickAction>
              </div>
            </div>

            <div className="animate-pop rounded-[20px] border border-hair bg-panel p-5 shadow-glass">
              <h3 className="m-0 mb-3 text-sm font-semibold text-fg">Workspace context</h3>
              <div className="flex flex-col gap-2.5 text-[13px]">
                <div className="flex items-center justify-between">
                  <span className="text-fg3">Workspace</span>
                  <span className="font-medium text-fg">{state?.workspaceName || state?.workspaceId}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-fg3">Project</span>
                  {state?.projectId ? (
                    <span className="rounded-full bg-pass-soft px-2.5 py-0.5 text-[11px] font-semibold text-pass">
                      {state.projectName || state.projectId.slice(0, 8)}
                    </span>
                  ) : (
                    <span className="rounded-full bg-warn-soft px-2.5 py-0.5 text-[11px] font-semibold text-warn">
                      No project
                    </span>
                  )}
                </div>
                {!state?.projectId && (
                  <p className="text-[11.5px] text-fg3">
                    Select a project in{" "}
                    <Link href="/dashboard/projects" className="text-acc hover:underline">
                      Projects
                    </Link>{" "}
                    to enable test case and run creation.
                  </p>
                )}
              </div>
            </div>
          </div>

          {!state?.projectId && <ProjectOnboarding onCreate={() => router.push("/dashboard/projects")} />}
        </>
      ) : (
        <div className="animate-pop rounded-[20px] border border-dashed border-hair-hi bg-panel p-10 text-center shadow-glass">
          <p className="text-[13px] text-fg2">Select a workspace to see quality metrics.</p>
        </div>
      )}
    </div>
  );
}

function QuickAction({
  href,
  icon: Icon,
  primary,
  children,
}: {
  href: string;
  icon: typeof Plus;
  primary?: boolean;
  children: React.ReactNode;
}) {
  return (
    <Link
      href={href}
      className={
        primary
          ? "flex h-9 items-center gap-2 rounded-xl bg-gradient-to-br from-acc to-acc2 px-3.5 text-[12.5px] font-semibold text-white transition-transform hover:-translate-y-px"
          : "flex h-9 items-center gap-2 rounded-xl border border-hair bg-panel px-3.5 text-[12.5px] font-medium text-fg2 transition-colors hover:bg-panel-hi hover:text-fg"
      }
    >
      <Icon className="h-3.5 w-3.5" aria-hidden="true" />
      {children}
    </Link>
  );
}
