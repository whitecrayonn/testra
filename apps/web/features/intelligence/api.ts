import { apiFetch } from "@/lib/api";
import type { FlakyPrediction, FailureCluster, FailureClassification, RunHistoryPoint } from "@/types/intelligence";

export async function listPredictions(workspaceId: string, minScore = 0.0, limit = 50): Promise<FlakyPrediction[]> {
  return apiFetch(`/api/v1/intelligence/predictions?workspace_id=${workspaceId}&min_score=${minScore}&limit=${limit}`);
}

export async function predictFlaky(input: {
  workspace_id: string;
  test_case_id: string;
  test_case_title: string;
  history: RunHistoryPoint[];
}): Promise<FlakyPrediction> {
  return apiFetch("/api/v1/intelligence/predictions", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function listClusters(workspaceId: string, limit = 50): Promise<FailureCluster[]> {
  return apiFetch(`/api/v1/intelligence/clusters?workspace_id=${workspaceId}&limit=${limit}`);
}

export async function classifyFailure(input: {
  workspace_id: string;
  error_message: string;
  stack_trace?: string;
}): Promise<FailureClassification> {
  return apiFetch("/api/v1/intelligence/classify", {
    method: "POST",
    body: JSON.stringify(input),
  });
}
