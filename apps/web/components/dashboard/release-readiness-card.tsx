"use client";

import { ShieldCheck, ShieldAlert } from "lucide-react";
import type { Metrics } from "@/types/analytics";

interface ReleaseReadinessCardProps {
  metrics: Metrics | null;
  loading: boolean;
}

const RING_RADIUS = 65;
const RING_CIRCUMFERENCE = 2 * Math.PI * RING_RADIUS;

export function ReleaseReadinessCard({ metrics, loading }: ReleaseReadinessCardProps) {
  const passRatePct = metrics ? Math.round(metrics.pass_rate * 100) : 0;
  const latest = metrics?.execution_timeline?.[metrics.execution_timeline.length - 1];
  const passed = latest?.passed ?? 0;
  const failed = latest?.failed ?? 0;
  const blocked = latest?.blocked ?? metrics?.blocked ?? 0;
  const avgSeconds = metrics ? (metrics.average_execution_time_ms / 1000).toFixed(2) : "0.00";
  const ready = !loading && (metrics?.open_defects ?? 0) === 0 && passRatePct >= 90;

  const offset = RING_CIRCUMFERENCE - (RING_CIRCUMFERENCE * Math.min(passRatePct, 100)) / 100;

  return (
    <div className="relative flex animate-pop gap-6 overflow-hidden rounded-[20px] border border-hair bg-panel p-5 shadow-glass">
      <div
        className="pointer-events-none absolute inset-0"
        style={{ background: "radial-gradient(420px 220px at 8% 0%, var(--acc-soft), transparent 70%)" }}
      />
      <div className="relative h-[152px] w-[152px] flex-none">
        <svg viewBox="0 0 152 152" width="152" height="152" className="block -rotate-90">
          <circle cx="76" cy="76" r={RING_RADIUS} fill="none" stroke="var(--hair)" strokeWidth="11" />
          <circle
            cx="76"
            cy="76"
            r={RING_RADIUS}
            fill="none"
            stroke="url(#readiness-ring-gradient)"
            strokeWidth="11"
            strokeLinecap="round"
            strokeDasharray={RING_CIRCUMFERENCE}
            strokeDashoffset={offset}
            style={{ transition: "stroke-dashoffset 1s cubic-bezier(.2,.8,.2,1)" }}
          />
          <defs>
            <linearGradient id="readiness-ring-gradient" x1="0" y1="0" x2="1" y2="1">
              <stop offset="0%" stopColor="var(--acc)" />
              <stop offset="100%" stopColor="var(--acc2)" />
            </linearGradient>
          </defs>
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center gap-px">
          <span className="font-mono text-[30px] font-bold tracking-tight text-fg">{passRatePct}%</span>
          <span className="font-mono text-[9px] tracking-[0.14em] text-fg3">PASS RATE</span>
        </div>
      </div>

      <div className="relative flex min-w-0 flex-1 flex-col justify-center gap-3.5">
        <div>
          <div
            className={
              "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[11.5px] font-bold tracking-wide " +
              (ready ? "bg-pass-soft text-pass" : "bg-warn-soft text-warn")
            }
          >
            {ready ? <ShieldCheck className="h-3.5 w-3.5" /> : <ShieldAlert className="h-3.5 w-3.5" />}
            {ready ? "READY TO SHIP" : "NEEDS ATTENTION"}
          </div>
          <h2 className="mt-2.5 text-[19px] font-semibold tracking-tight text-fg">Release readiness</h2>
          <p className="mt-1 max-w-[380px] text-[12.5px] text-fg2">
            {metrics
              ? `${metrics.open_defects} open defect${metrics.open_defects === 1 ? "" : "s"} across ${metrics.total_test_cases} test cases, ${metrics.total_test_runs} runs recorded.`
              : "No execution data yet for this workspace."}
          </p>
        </div>
        <div className="flex gap-5.5">
          <Stat label="PASSED" value={passed} color="text-pass" />
          <Stat label="FAILED" value={failed} color="text-fail" />
          <Stat label="BLOCKED" value={blocked} color="text-warn" />
          <Stat label="AVG" value={`${avgSeconds}s`} color="text-fg2" />
        </div>
      </div>
    </div>
  );
}

function Stat({ label, value, color }: { label: string; value: number | string; color: string }) {
  return (
    <div>
      <div className={`font-mono text-[17px] font-bold ${color}`}>{value}</div>
      <div className="text-[10.5px] tracking-wide text-fg3">{label}</div>
    </div>
  );
}
