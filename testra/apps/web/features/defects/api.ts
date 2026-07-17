import { apiFetch } from "@/lib/api";
import type { Defect, PaginationMeta } from "@/types/defects";

interface PaginatedResponse<T> {
  data: T[];
  meta: PaginationMeta;
}

export async function listDefects(
  projectId: string,
  params?: { cursor?: string; limit?: number },
): Promise<{ data: Defect[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams({ project_id: projectId });
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/defects?${searchParams.toString()}`) as Promise<
    PaginatedResponse<Defect>
  >;
}

export async function getDefect(id: string): Promise<Defect> {
  return apiFetch(`/api/v1/defects/${id}`);
}

export async function createDefect(input: {
  workspace_id: string;
  project_id: string;
  test_run_item_id?: string;
  title: string;
  description?: string;
  severity?: string;
  priority?: string;
  status?: string;
  assigned_to?: string;
}): Promise<Defect> {
  return apiFetch("/api/v1/defects", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateDefect(
  id: string,
  input: {
    title?: string;
    description?: string;
    severity?: string;
    priority?: string;
    status?: string;
    assigned_to?: string | null;
    test_run_item_id?: string | null;
  },
): Promise<Defect> {
  return apiFetch(`/api/v1/defects/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function deleteDefect(id: string): Promise<void> {
  await apiFetch(`/api/v1/defects/${id}`, { method: "DELETE" });
}
