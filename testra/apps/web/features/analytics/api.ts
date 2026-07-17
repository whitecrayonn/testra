import { apiFetch } from "@/lib/api";
import type { Dashboard, Summary, TrendPoint } from "@/types/analytics";

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
