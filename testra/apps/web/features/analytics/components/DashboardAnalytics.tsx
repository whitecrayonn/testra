"use client";

import { useEffect, useMemo, useState } from "react";
import { Download, Activity as ActivityIcon, Loader2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { PageHeader } from "@/components/ui/page-header";
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
  title = "Dashboard",
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
    () => (metrics?.execution_timeline ?? []).slice().reverse().map((p) => ({ ...p, date: p.date.slice(5) })),
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
        name: it.title || it.test_case_id.slice(0, 8),
        failures: it.failures,
      })),
    [metrics?.top_failed_test_cases],
  );

  const topApiData = useMemo(
    () =>
      (metrics?.top_failed_apis ?? []).map((it) => ({
        name: it.name || it.request_id.slice(0, 8),
        failures: it.failures,
      })),
    [metrics?.top_failed_apis],
  );

  const activeQaData = useMemo(
    () =>
      (metrics?.most_active_qa ?? []).map((u) => ({
        name: u.name || u.user_id.slice(0, 8),
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
      <div className="flex h-96 items-center justify-center text-slate-500">
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
    <div className="space-y-6">
      <PageHeader
        title={title}
        description={description}
        actions={
          <a
            href={getMetricsCSVUrl(filter)}
            download
            className="inline-flex items-center rounded-md border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100 dark:hover:bg-slate-700"
          >
            <Download className="mr-2 h-4 w-4" aria-hidden="true" />
            Export CSV
          </a>
        }
      />

      <DashboardFilters value={filter} onChange={setFilter} />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
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

      <div className="grid gap-6 lg:grid-cols-2">
        <ChartCard title="Execution Timeline" data={trendData.length ? trendData : placeholderTrend()}>
          <LineChartComponent data={trendData.length ? trendData : placeholderTrend()} xKey="date" yKey="total_runs" className="h-64" />
        </ChartCard>

        <ChartCard title="Weekly Trend" data={metrics.weekly_trend}>
          <StackedBarChart
            data={stackTrend(metrics.weekly_trend)}
            xKey="date"
            keys={["passed", "failed", "skipped", "blocked"]}
            colors={["#22c55e", "#ef4444", "#eab308", "#f97316"]}
            className="h-64"
          />
        </ChartCard>

        <ChartCard title="Release Quality" data={releaseData}>
          <StackedBarChart
            data={releaseData}
            xKey="release"
            keys={["passed", "failed", "skipped", "blocked"]}
            colors={["#22c55e", "#ef4444", "#eab308", "#f97316"]}
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
              {activity.slice(0, 20).map((a) => (
                <div key={a.id + a.type} className="flex items-center justify-between rounded-lg border border-slate-100 p-2 text-sm dark:border-slate-700">
                  <span className="font-medium text-slate-900 dark:text-slate-100">{a.title}</span>
                  <span className="text-xs capitalize text-slate-500 dark:text-slate-400">{a.type}</span>
                </div>
              ))}
              {activity.length === 0 && <p className="text-sm text-slate-500">No recent activity.</p>}
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
    success: "text-green-600",
    danger: "text-red-600",
    warning: "text-orange-600",
    info: "text-brand-600",
    default: "text-slate-900",
  };
  const color = colorMap[variant || "default"];
  return (
    <Card className="p-4 dark:border-slate-700 dark:bg-slate-900">
      <p className="text-xs text-slate-500 dark:text-slate-400">{label}</p>
      <p className={`text-xl font-bold ${color} dark:text-opacity-90`}>{value}</p>
    </Card>
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
    <Card className="dark:border-slate-700 dark:bg-slate-900">
      <CardHeader>
        <CardTitle className="text-sm font-medium text-slate-900 dark:text-slate-100">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {data.length === 0 ? (
          <div className="flex h-64 items-center justify-center text-sm text-slate-500">No data available.</div>
        ) : (
          children
        )}
      </CardContent>
    </Card>
  );
}

function stackTrend(trends: TrendPoint[]) {
  return trends
    .slice()
    .reverse()
    .map((p) => ({ ...p, date: p.date.slice(5) }));
}

function placeholderTrend(): TrendPoint[] {
  return [];
}
