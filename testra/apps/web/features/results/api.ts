import { apiFetch } from "@/lib/api";
import type { TestRun, TestRunItem, PaginationMeta } from "@/types/results";

interface PaginatedResponse<T> {
  data: T[];
  meta: PaginationMeta;
}

export async function listTestRuns(
  projectId: string,
  params?: { cursor?: string; limit?: number },
): Promise<{ data: TestRun[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams({ project_id: projectId });
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/test-runs?${searchParams.toString()}`) as Promise<
    PaginatedResponse<TestRun>
  >;
}

export async function getTestRun(id: string): Promise<TestRun> {
  return apiFetch(`/api/v1/test-runs/${id}`);
}

export async function createTestRun(input: {
  workspace_id: string;
  project_id: string;
  suite_id?: string;
  name: string;
  test_case_ids?: string[];
  source?: string;
}): Promise<TestRun> {
  return apiFetch("/api/v1/test-runs", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateTestRunStatus(
  id: string,
  status: string,
): Promise<TestRun> {
  return apiFetch(`/api/v1/test-runs/${id}`, {
    method: "PUT",
    body: JSON.stringify({ status }),
  });
}

export async function deleteTestRun(id: string): Promise<void> {
  await apiFetch(`/api/v1/test-runs/${id}`, { method: "DELETE" });
}

export async function listTestRunItems(runId: string): Promise<TestRunItem[]> {
  const result = await apiFetch(`/api/v1/test-runs/${runId}/items`);
  if (Array.isArray(result)) return result;
  return (result as { data: TestRunItem[] }).data;
}

export async function updateTestRunItemStatus(
  id: string,
  input: {
    status: string;
    duration_ms?: number;
    error_message?: string;
    stack_trace?: string;
  },
): Promise<TestRunItem> {
  return apiFetch(`/api/v1/test-run-items/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}
