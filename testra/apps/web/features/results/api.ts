import { apiFetch } from "@/lib/api";
import type {
  TestRun,
  TestRunItem,
  TestPlan,
  TestPlanItem,
  StepResult,
  Evidence,
  ExecutionHistoryEntry,
  PaginationMeta,
  RunItemStatus,
} from "@/types/results";

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

export async function cloneTestRun(id: string): Promise<TestRun> {
  return apiFetch(`/api/v1/test-runs/${id}/clone`, { method: "POST" });
}

export async function rerunTestRun(id: string): Promise<TestRun> {
  return apiFetch(`/api/v1/test-runs/${id}/rerun`, { method: "POST" });
}

export async function bulkUpdateTestRunItems(
  runId: string,
  input: { item_ids: string[]; status: RunItemStatus },
): Promise<{ data: TestRunItem[] }> {
  return apiFetch(`/api/v1/test-runs/${runId}/bulk`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function listTestRunItems(
  runId: string,
  params?: { status?: string; search?: string; cursor?: string; limit?: number },
): Promise<{ data: TestRunItem[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams();
  if (params?.status) searchParams.set("status", params.status);
  if (params?.search) searchParams.set("search", params.search);
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  const query = searchParams.toString();
  return apiFetch(`/api/v1/test-runs/${runId}/items${query ? `?${query}` : ""}`);
}

export async function updateTestRunItemStatus(
  id: string,
  input: {
    status: RunItemStatus;
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

export async function executeTestRunItem(
  id: string,
  input: {
    status: RunItemStatus;
    step_results?: StepResult[];
    comment?: string;
    duration_ms?: number;
    error_message?: string;
    stack_trace?: string;
  },
): Promise<TestRunItem> {
  return apiFetch(`/api/v1/test-run-items/${id}/execute`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function getExecutionHistory(
  itemId: string,
): Promise<{ data: ExecutionHistoryEntry[] }> {
  return apiFetch(`/api/v1/test-run-items/${itemId}/history`);
}

export async function attachEvidence(
  itemId: string,
  input: {
    step_order?: number;
    file_name: string;
    content_type?: string;
    storage_path?: string;
  },
): Promise<{ data: Evidence }> {
  return apiFetch(`/api/v1/test-run-items/${itemId}/evidence`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function listEvidence(itemId: string): Promise<{ data: Evidence[] }> {
  return apiFetch(`/api/v1/test-run-items/${itemId}/evidence`);
}

export async function deleteEvidence(itemId: string, evidenceId: string): Promise<void> {
  await apiFetch(`/api/v1/test-run-items/${itemId}/evidence/${evidenceId}`, {
    method: "DELETE",
  });
}

export async function linkDefect(
  itemId: string,
  defectId: string,
): Promise<void> {
  await apiFetch(`/api/v1/test-run-items/${itemId}/defects`, {
    method: "POST",
    body: JSON.stringify({ defect_id: defectId }),
  });
}

export async function listLinkedDefects(itemId: string): Promise<{ data: string[] }> {
  return apiFetch(`/api/v1/test-run-items/${itemId}/defects`);
}

export async function unlinkDefect(
  itemId: string,
  defectId: string,
): Promise<void> {
  await apiFetch(`/api/v1/test-run-items/${itemId}/defects/${defectId}`, {
    method: "DELETE",
  });
}

export async function listTestPlans(
  projectId: string,
  params?: { cursor?: string; limit?: number },
): Promise<{ data: TestPlan[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams({ project_id: projectId });
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/test-plans?${searchParams.toString()}`);
}

export async function getTestPlan(id: string): Promise<{ data: TestPlan }> {
  return apiFetch(`/api/v1/test-plans/${id}`);
}

export async function createTestPlan(input: {
  workspace_id: string;
  project_id: string;
  suite_id?: string;
  name: string;
  description?: string;
  configuration?: Record<string, unknown>;
  test_case_ids?: string[];
}): Promise<{ data: TestPlan }> {
  return apiFetch("/api/v1/test-plans", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateTestPlan(
  id: string,
  input: Partial<{
    name: string;
    description: string;
    status: "active" | "archived";
    configuration: Record<string, unknown>;
    test_case_ids: string[];
  }>,
): Promise<{ data: TestPlan }> {
  return apiFetch(`/api/v1/test-plans/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function deleteTestPlan(id: string): Promise<void> {
  await apiFetch(`/api/v1/test-plans/${id}`, { method: "DELETE" });
}

export async function getTestPlanItems(
  planId: string,
): Promise<{ data: TestPlanItem[] }> {
  return apiFetch(`/api/v1/test-plans/${planId}/items`);
}

export async function createRunFromPlan(planId: string): Promise<{ data: TestRun }> {
  return apiFetch(`/api/v1/test-plans/${planId}/runs`, { method: "POST" });
}
