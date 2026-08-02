"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { Plus, ChevronRight, Play, CheckCircle2, XCircle, Clock, SkipForward, Ban } from "lucide-react";
import { StatusPill, type StatusTone } from "@/components/ui/status-pill";
import { SegmentedBar } from "@/components/ui/segmented-bar";
import { listTestRuns } from "@/features/results/api";
import type { TestRun, PaginationMeta } from "@/types/results";

const STATUS_TONE: Record<TestRun["status"], StatusTone> = {
  pending: "neutral",
  running: "info",
  passed: "pass",
  failed: "fail",
  skipped: "warn",
  cancelled: "neutral",
};

const STATUS_ICON: Record<TestRun["status"], typeof Play> = {
  pending: Clock,
  running: Play,
  passed: CheckCircle2,
  failed: XCircle,
  skipped: SkipForward,
  cancelled: Ban,
};

export default function TestRunsPage() {
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [cursor, setCursor] = useState<string | undefined>(undefined);
  const [error, setError] = useState<string | null>(null);

  const projectId =
    typeof window !== "undefined"
      ? localStorage.getItem("testra_project_id") || ""
      : "";

  const fetchRuns = useCallback(
    async (reset?: boolean) => {
      setLoading(true);
      setError(null);
      try {
        const result = await listTestRuns(projectId, {
          cursor: reset ? undefined : cursor,
        });
        setRuns((prev) => (reset ? result.data : [...prev, ...result.data]));
        setMeta(result.meta);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load runs");
      } finally {
        setLoading(false);
      }
    },
    [projectId, cursor],
  );

  useEffect(() => {
    if (projectId) {
      fetchRuns(true);
    } else {
      setLoading(false);
    }
  }, [projectId, fetchRuns]);

  const loadMore = () => {
    if (meta?.next_cursor) {
      setCursor(meta.next_cursor);
    }
  };

  useEffect(() => {
    if (cursor) {
      fetchRuns(false);
    }
  }, [cursor, fetchRuns]);

  return (
    <div className="flex flex-col gap-3.5">
      <div className="flex animate-rise-sm items-end justify-between gap-4">
        <div>
          <h1 className="m-0 text-[27px] font-bold tracking-tight text-fg">Test Runs</h1>
          <p className="mt-1.5 text-[13px] text-fg2">View and manage test execution runs.</p>
        </div>
        <Link
          href="/dashboard/test-runs/new"
          className="flex h-[38px] items-center gap-2 rounded-[13px] bg-gradient-to-br from-acc to-acc2 px-4 text-[13px] font-bold text-white shadow-[0_10px_26px_-12px_var(--ring)] transition-transform hover:-translate-y-px"
        >
          <Plus className="h-3.5 w-3.5" aria-hidden="true" />
          New Run
        </Link>
      </div>

      {error && (
        <div role="alert" className="rounded-[16px] border border-fail-soft bg-fail-soft p-4 text-[13px] text-fail">
          {error}
        </div>
      )}

      {!projectId ? (
        <NoProject />
      ) : loading && runs.length === 0 ? (
        <ListSkeleton count={3} />
      ) : runs.length === 0 ? (
        <div className="flex animate-pop flex-col items-center justify-center gap-2 rounded-[22px] border border-dashed border-hair-hi bg-panel p-16 text-center shadow-glass">
          <div className="mb-1.5 flex h-[62px] w-[62px] animate-float-y items-center justify-center rounded-[20px] bg-acc-soft text-acc">
            <Play className="h-6 w-6" aria-hidden="true" />
          </div>
          <h3 className="m-0 text-[19px] font-semibold tracking-tight text-fg">No test runs yet</h3>
          <p className="m-0 max-w-sm text-[13px] text-fg2">Create your first test run to track execution progress.</p>
          <div className="mt-3.5 flex gap-2.5">
            <Link
              href="/dashboard/test-runs/new"
              className="flex h-[38px] items-center gap-2 rounded-[13px] bg-gradient-to-br from-acc to-acc2 px-4 text-[13px] font-bold text-white"
            >
              <Plus className="h-3.5 w-3.5" aria-hidden="true" />
              New Run
            </Link>
            <Link
              href="/dashboard/projects"
              className="flex h-[38px] items-center rounded-[13px] border border-hair-hi bg-panel px-4 text-[13px] font-semibold text-fg transition-colors hover:bg-panel-hi"
            >
              Go to Projects
            </Link>
          </div>
        </div>
      ) : (
        <div className="flex flex-col gap-2.5">
          {runs.map((run, i) => {
            const Icon = STATUS_ICON[run.status];
            const tone = STATUS_TONE[run.status];
            return (
              <Link
                key={run.id}
                href={`/dashboard/test-runs/${run.id}`}
                style={{ animationDelay: `${Math.min(i, 8) * 40}ms` }}
                className="flex animate-rise-sm items-center gap-4 rounded-[18px] border border-hair bg-panel p-4 shadow-glass transition-all hover:-translate-y-0.5 hover:border-hair-hi hover:bg-panel-hi sm:gap-5"
              >
                <span className="flex h-[38px] w-[38px] flex-none items-center justify-center rounded-[13px] bg-acc-soft text-acc">
                  <Icon className="h-4 w-4" aria-hidden="true" />
                </span>
                <div className="min-w-0 flex-1">
                  <h3 className="m-0 text-[14.5px] font-semibold text-fg">{run.name}</h3>
                  <div className="mt-1 font-mono text-[11px] text-fg3">
                    {run.source} · {run.passed}P · {run.failed}F · {run.skipped}S · {run.total} total
                  </div>
                </div>
                <div className="hidden w-[190px] flex-none flex-col gap-1.5 sm:flex">
                  <SegmentedBar passed={run.passed} failed={run.failed} skipped={run.skipped} blocked={run.blocked} total={run.total} />
                </div>
                <StatusPill tone={tone}>{run.status}</StatusPill>
                <span className="hidden w-[58px] flex-none text-right font-mono text-[11px] text-fg3 md:inline">
                  {new Date(run.created_at).toLocaleDateString()}
                </span>
                <ChevronRight className="h-4 w-4 flex-none text-fg3" aria-hidden="true" />
              </Link>
            );
          })}

          {meta?.has_more && (
            <div className="flex justify-center pt-3.5">
              <button
                onClick={loadMore}
                disabled={loading}
                className="h-9 rounded-[13px] border border-hair-hi bg-panel px-4 text-[12.5px] font-semibold text-fg2 transition-colors hover:bg-panel-hi disabled:opacity-50"
              >
                {loading ? "Loading…" : "Load More"}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function NoProject() {
  return (
    <div className="flex animate-pop flex-col items-center justify-center gap-2 rounded-[22px] border border-dashed border-hair-hi bg-panel p-16 text-center shadow-glass">
      <div className="mb-1.5 flex h-[62px] w-[62px] items-center justify-center rounded-[20px] bg-acc-soft text-acc">
        <Play className="h-6 w-6" aria-hidden="true" />
      </div>
      <h3 className="m-0 text-[19px] font-semibold tracking-tight text-fg">No project selected</h3>
      <p className="m-0 max-w-sm text-[13px] text-fg2">Select a project from the Projects page to view and create test runs.</p>
      <Link
        href="/dashboard/projects"
        className="mt-3.5 flex h-[38px] items-center rounded-[13px] bg-gradient-to-br from-acc to-acc2 px-4 text-[13px] font-bold text-white"
      >
        Go to Projects
      </Link>
    </div>
  );
}

function ListSkeleton({ count }: { count: number }) {
  return (
    <div className="flex flex-col gap-2.5">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="h-[70px] animate-pulse rounded-[18px] border border-hair bg-panel" />
      ))}
    </div>
  );
}
