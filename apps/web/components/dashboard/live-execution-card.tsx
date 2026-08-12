"use client";

import { Activity } from "lucide-react";
import type { TestRun } from "@/types/results";

interface LiveExecutionCardProps {
  runs: TestRun[];
  loading: boolean;
}

const SOURCE_LABEL: Record<TestRun["source"], string> = {
  manual: "Manual run",
  ci: "CI",
  api: "API",
};

export function LiveExecutionCard({ runs, loading }: LiveExecutionCardProps) {
  const running = runs.find((r) => r.status === "running");

  if (loading) {
    return (
      <div className="flex animate-pop flex-col justify-center gap-2 rounded-[20px] border border-hair bg-panel p-5 shadow-glass">
        <div className="h-4 w-24 animate-pulse rounded bg-hair" />
        <div className="h-6 w-40 animate-pulse rounded bg-hair" />
      </div>
    );
  }

  if (!running) {
    return (
      <div className="flex animate-pop flex-col items-center justify-center gap-2 rounded-[20px] border border-hair bg-panel p-5 text-center shadow-glass">
        <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-acc-soft text-acc">
          <Activity className="h-5 w-5" aria-hidden="true" />
        </div>
        <div className="text-[13px] font-medium text-fg2">No run in progress</div>
        <p className="text-[11.5px] text-fg3">Executions will appear here live while they run.</p>
      </div>
    );
  }

  const completed = running.passed + running.failed + running.skipped + running.blocked;
  const pct = running.total > 0 ? Math.round((completed / running.total) * 100) : 0;

  return (
    <div className="flex animate-pop flex-col gap-3.5 overflow-hidden rounded-[20px] border border-hair bg-panel p-5 shadow-glass">
      <div className="flex items-center gap-2">
        <span className="h-1.5 w-1.5 animate-dot-pulse rounded-full bg-info" />
        <span className="font-mono text-[10px] tracking-[0.14em] text-fg3">LIVE EXECUTION</span>
      </div>
      <div>
        <div className="text-[16px] font-semibold tracking-tight text-fg">{running.name}</div>
        <div className="mt-0.5 text-[11.5px] text-fg3">{SOURCE_LABEL[running.source]} · started {new Date(running.started_at ?? running.created_at).toLocaleTimeString()}</div>
      </div>
      <div>
        <div className="relative h-2 overflow-hidden rounded-full bg-hair">
          <div
            className="absolute inset-y-0 left-0 rounded-full bg-gradient-to-r from-acc to-acc2"
            style={{ width: `${pct}%`, transition: "width .5s cubic-bezier(.2,.8,.2,1)" }}
          />
        </div>
        <div className="mt-2 flex justify-between font-mono text-[11px] text-fg3">
          <span>{completed} / {running.total} items</span>
          <span className="text-fg2">{pct}%</span>
        </div>
      </div>
      <div className="mt-0.5 flex gap-3 text-[11px] text-fg2">
        <span className="text-pass">{running.passed} passed</span>
        <span className="text-fail">{running.failed} failed</span>
        {running.skipped > 0 && <span className="text-warn">{running.skipped} skipped</span>}
      </div>
    </div>
  );
}
