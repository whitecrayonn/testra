"use client";

import { useEffect, useMemo, useState } from "react";
import { Download, Activity as ActivityIcon, Loader2 } from "lucide-react";
import { EmptyState } from "@/components/ui/empty-state";
import { DashboardFilters } from "@/components/dashboard/filters";
import {
  BarChartComponent,
  LineChartComponent,
  PieChartComponent,
  StackedBarChart,
  chartColors,
} from "@/components/charts";
import { getActivity, getMetrics, getMetricsCSVUrl } from "@/features/analytics/api";
import type { Activity, Metrics, MetricsFilter, TrendPoint } from "@/types/analytics";

interface DashboardAnalyticsProps {
  workspaceId: string;
  projectId?: string;
  title?: string;
  description?: string;
  source?: string;
  tester?: string;
}

export function DashboardAnalytics({
  workspaceId,
  projectId,
  title = "Analytics Overview",
  description = "Executive overview of test execution, defects, and team activity.",
  source,
  tester,
}: DashboardAnalyticsProps) {
  const [filter, setFilter] = useState<MetricsFilter>({
    workspace_id: workspaceId,
    project_id: projectId,
    source,
    tester,
  });
  const [metrics, setMetrics] = useState<Metrics | null>(null);
  const [activity, setActivity] = useState<Activity[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    Promise.all([getMetrics(filter), getActivity(filter)])
      .then(([m, a]) => {
        setMetrics(m);
        setActivity(a);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load analytics"))
      .finally(() => setLoading(false));
  }, [filter]);

  const trendData = useMemo(
    () => (metrics?.execution_timeline ?? []).slice().reverse().map((p) => ({ ...p, date: p.date?.slice(5) ?? p.date })),
    [metrics?.execution_timeline],
  );

  const releaseData = useMemo(
    () =>
      (metrics?.release_quality_trend ?? []).map((p) => ({
        release: p.release || "N/A",
        passed: p.passed,
        failed: p.failed,
        skipped: p.skipped,
        blocked: p.blocked,
      })),
    [metrics?.release_quality_trend],
  );

  const topFailedData = useMemo(
    () =>
      (metrics?.top_failed_test_cases ?? []).map((it) => ({
        name: it.title || it.test_case_id?.slice(0, 8) || "Unknown",
        failures: it.failures,
      })),
    [metrics?.top_failed_test_cases],
  );

  const topApiData = useMemo(
    () =>
      (metrics?.top_failed_apis ?? []).map((it) => ({
        name: it.name || it.request_id?.slice(0, 8) || "Unknown",
        failures: it.failures,
      })),
    [metrics?.top_failed_apis],
  );

  const activeQaData = useMemo(
    () =>
      (metrics?.most_active_qa ?? []).map((u) => ({
        name: u.name || u.user_id?.slice(0, 8) || "Unknown",
        executions: u.count,
      })),
    [metrics?.most_active_qa],
  );

  const coverageData = useMemo(
    () => [
      { name: "Automation", value: metrics?.automation_coverage || 0 },
      { name: "API", value: metrics?.api_test_coverage || 0 },
      { name: "Manual", value: 100 - (metrics?.automation_coverage || 0) - (metrics?.api_test_coverage || 0) },
    ],
    [metrics?.automation_coverage, metrics?.api_test_coverage],
  );

  if (loading && !metrics) {
    return (
      <div className="flex h-96 items-center justify-center text-fg3">
        <Loader2 className="mr-2 h-5 w-5 animate-spin" aria-hidden="true" />
        Loading analytics...
      </div>
    );
  }

  if (error) {
    return <EmptyState icon={ActivityIcon} title="Analytics error" description={error} />;
  }

  if (!metrics) {
    return <EmptyState icon={ActivityIcon} title="No data" description="No analytics data available." />;
  }

  return (
    <div className="flex flex-col gap-3.5">
      <div className="flex animate-rise-sm items-end justify-between gap-4">
        <div>
          <h1 className="m-0 text-[21px] font-bold tracking-tight text-fg">{title}</h1>
          <p className="mt-1 text-[13px] text-fg2">{description}</p>
        </div>
        <a
          href={getMetricsCSVUrl(filter)}
          download
          className="flex h-9 items-center gap-2 rounded-[13px] border border-hair-hi bg-panel px-3.5 text-[12.5px] font-semibold text-fg transition-colors hover:bg-panel-hi"
        >
          <Download className="h-3.5 w-3.5" aria-hidden="true" />
          Export CSV
        </a>
      </div>

      <DashboardFilters value={filter} onChange={setFilter} />

      <div className="grid gap-3.5 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard label="Total Test Cases" value={metrics.total_test_cases} />
        <MetricCard label="Total Test Plans" value={metrics.total_test_plans} />
        <MetricCard label="Total Test Runs" value={metrics.total_test_runs} />
        <MetricCard label="Execution Progress" value={`${(metrics.execution_progress * 100).toFixed(1)}%`} />
        <MetricCard label="Pass Rate" value={`${(metrics.pass_rate * 100).toFixed(1)}%`} variant="success" />
        <MetricCard label="Fail Rate" value={`${(metrics.fail_rate * 100).toFixed(1)}%`} variant="danger" />
        <MetricCard label="Blocked" value={metrics.blocked} variant="warning" />
        <MetricCard label="Retest" value={metrics.retest} variant="info" />
        <MetricCard label="Skipped" value={metrics.skipped} />
        <MetricCard label="Automation Coverage" value={`${metrics.automation_coverage.toFixed(1)}%`} />
        <MetricCard label="API Test Coverage" value={`${metrics.api_test_coverage.toFixed(1)}%`} />
        <MetricCard label="Avg Exec Time" value={`${(metrics.average_execution_time_ms / 1000).toFixed(2)}s`} />
        <MetricCard label="Defect Density" value={`${metrics.defect_density.toFixed(2)}%`} />
        <MetricCard label="Open Defects" value={metrics.open_defects} variant="danger" />
        <MetricCard label="Closed Defects" value={metrics.closed_defects} variant="success" />
        <MetricCard label="Bug Reopen Rate" value={`${metrics.bug_reopen_rate.toFixed(1)}%`} />
      </div>

      <div className="grid gap-3.5 lg:grid-cols-2">
        <ChartCard title="Execution Timeline" data={trendData.length ? trendData : placeholderTrend()}>
          <LineChartComponent data={trendData.length ? trendData : placeholderTrend()} xKey="date" yKey="total_runs" className="h-64" />
        </ChartCard>

        <ChartCard title="Weekly Trend" data={metrics.weekly_trend ?? []}>
          <StackedBarChart
            data={stackTrend(metrics.weekly_trend)}
            xKey="date"
            keys={["passed", "failed", "skipped", "blocked"]}
            colors={["var(--pass)", "var(--fail)", "var(--warn)", "var(--info)"]}
            className="h-64"
          />
        </ChartCard>

        <ChartCard title="Release Quality" data={releaseData}>
          <StackedBarChart
            data={releaseData}
            xKey="release"
            keys={["passed", "failed", "skipped", "blocked"]}
            colors={["var(--pass)", "var(--fail)", "var(--warn)", "var(--info)"]}
            className="h-64"
          />
        </ChartCard>

        <ChartCard title="Coverage" data={coverageData}>
          <PieChartComponent data={coverageData} xKey="name" nameKey="name" dataKey="value" colors={chartColors} className="h-64" />
        </ChartCard>

        <ChartCard title="Top Failed Test Cases" data={topFailedData}>
          <BarChartComponent data={topFailedData} xKey="name" yKey="failures" className="h-64" />
        </ChartCard>

        <ChartCard title="Top Failed APIs" data={topApiData}>
          <BarChartComponent data={topApiData} xKey="name" yKey="failures" className="h-64" />
        </ChartCard>

        <ChartCard title="Most Active QA" data={activeQaData}>
          <BarChartComponent data={activeQaData} xKey="name" yKey="executions" className="h-64" />
        </ChartCard>

        <ChartCard title="Recent Activity" data={activity}>
          <div className="h-64 overflow-y-auto pr-2">
            <div className="space-y-2">
              {(activity ?? []).slice(0, 20).map((a) => (
                <div key={a.id + a.type} className="flex items-center justify-between rounded-[13px] border border-hair bg-panel p-2.5 text-[12.5px]">
                  <span className="font-medium text-fg">{a.title}</span>
                  <span className="text-[11px] capitalize text-fg3">{a.type}</span>
                </div>
              ))}
              {(activity ?? []).length === 0 && <p className="text-[12.5px] text-fg3">No recent activity.</p>}
            </div>
          </div>
        </ChartCard>
      </div>
    </div>
  );
}

function MetricCard({
  label,
  value,
  variant,
}: {
  label: string;
  value: number | string;
  variant?: "success" | "danger" | "warning" | "info";
}) {
  const colorMap: Record<string, string> = {
    success: "text-pass",
    danger: "text-fail",
    warning: "text-warn",
    info: "text-info",
    default: "text-fg",
  };
  const color = colorMap[variant || "default"];
  return (
    <div className="rounded-2xl border border-hair bg-panel p-4 shadow-glass transition-transform hover:-translate-y-0.5">
      <p className="text-[11px] text-fg3">{label}</p>
      <p className={`mt-1 font-mono text-xl font-bold ${color}`}>{value}</p>
    </div>
  );
}

function ChartCard({
  title,
  data,
  children,
}: {
  title: string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data: any[];
  children: React.ReactNode;
}) {
  return (
    <div className="animate-pop rounded-[20px] border border-hair bg-panel p-4 shadow-glass">
      <h3 className="m-0 mb-3 text-sm font-semibold text-fg">{title}</h3>
      {data.length === 0 ? (
        <div className="flex h-64 items-center justify-center text-[13px] text-fg3">No data available.</div>
      ) : (
        children
      )}
    </div>
  );
}

function stackTrend(trends: TrendPoint[] | null | undefined) {
  return (trends ?? [])
    .slice()
    .reverse()
    .map((p) => ({ ...p, date: p.date?.slice(5) ?? "" }));
}

function placeholderTrend(): TrendPoint[] {
  return [];
}
