import { apiFetch } from "@/lib/api";
import type {
  Activity,
  Dashboard,
  Metrics,
  MetricsFilter,
  Summary,
  TrendPoint,
} from "@/types/analytics";

function buildParams(filter: MetricsFilter): URLSearchParams {
  const params = new URLSearchParams({ workspace_id: filter.workspace_id });
  if (filter.project_id) params.set("project_id", filter.project_id);
  if (filter.release) params.set("release", filter.release);
  if (filter.sprint) params.set("sprint", filter.sprint);
  if (filter.environment) params.set("environment", filter.environment);
  if (filter.tester) params.set("tester", filter.tester);
  if (filter.source) params.set("source", filter.source);
  if (filter.start) params.set("start", filter.start);
  if (filter.end) params.set("end", filter.end);
  return params;
}

export async function getSummary(workspaceId: string, projectId?: string): Promise<Summary> {
  const params = new URLSearchParams({ workspace_id: workspaceId });
  if (projectId) params.set("project_id", projectId);
  return apiFetch(`/api/v1/analytics/summary?${params.toString()}`);
}

export async function getTrends(
  workspaceId: string,
  projectId?: string,
  start?: string,
  end?: string,
): Promise<TrendPoint[]> {
  const params = new URLSearchParams({ workspace_id: workspaceId });
  if (projectId) params.set("project_id", projectId);
  if (start) params.set("start", start);
  if (end) params.set("end", end);
  return apiFetch(`/api/v1/analytics/trends?${params.toString()}`);
}

export async function getMetrics(filter: MetricsFilter): Promise<Metrics> {
  return apiFetch(`/api/v1/analytics/metrics?${buildParams(filter).toString()}`);
}

export async function getActivity(filter: MetricsFilter): Promise<Activity[]> {
  return apiFetch(`/api/v1/analytics/activity?${buildParams(filter).toString()}`);
}

export function getMetricsCSVUrl(filter: MetricsFilter): string {
  return `/api/v1/analytics/export/csv?${buildParams(filter).toString()}`;
}

export async function listDashboards(workspaceId: string): Promise<Dashboard[]> {
  return apiFetch(`/api/v1/analytics/dashboards?workspace_id=${workspaceId}`);
}

export async function createDashboard(input: {
  workspace_id: string;
  name: string;
  type?: string;
  config?: Record<string, unknown>;
}): Promise<Dashboard> {
  return apiFetch("/api/v1/analytics/dashboards", {
    method: "POST",
    body: JSON.stringify(input),
  });
}
