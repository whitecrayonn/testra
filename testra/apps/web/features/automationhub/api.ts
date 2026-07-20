import { apiFetch } from "@/lib/api";
import type {
  AutomationProject,
  AutomationExecution,
  AutomationArtifact,
  AutomationLog,
  IngestResult,
  PaginationMeta,
} from "@/types/automationhub";

interface PaginatedResponse<T> {
  data: T[];
  meta: PaginationMeta;
}

export async function listAutomationProjects(
  workspaceId: string,
  params?: { cursor?: string; limit?: number },
): Promise<{ data: AutomationProject[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams({ workspace_id: workspaceId });
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/automation/projects?${searchParams.toString()}`) as Promise<
    PaginatedResponse<AutomationProject>
  >;
}

export async function createAutomationProject(input: {
  workspace_id: string;
  project_id?: string;
  name: string;
  framework: string;
  repository_url?: string;
  branch?: string;
  command?: string;
}): Promise<AutomationProject> {
  return apiFetch("/api/v1/automation/projects", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function getAutomationProject(id: string): Promise<AutomationProject> {
  return apiFetch(`/api/v1/automation/projects/${id}`);
}

export async function updateAutomationProject(
  id: string,
  input: {
    name: string;
    framework: string;
    repository_url?: string;
    branch?: string;
    command?: string;
  },
): Promise<AutomationProject> {
  return apiFetch(`/api/v1/automation/projects/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function deleteAutomationProject(id: string): Promise<void> {
  await apiFetch(`/api/v1/automation/projects/${id}`, { method: "DELETE" });
}

export async function listAutomationExecutions(
  projectId: string,
  params?: { cursor?: string; limit?: number },
): Promise<{ data: AutomationExecution[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams({ project_id: projectId });
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/automation/executions?${searchParams.toString()}`) as Promise<
    PaginatedResponse<AutomationExecution>
  >;
}

export async function getAutomationExecution(id: string): Promise<AutomationExecution> {
  return apiFetch(`/api/v1/automation/executions/${id}`);
}

export async function deleteAutomationExecution(id: string): Promise<void> {
  await apiFetch(`/api/v1/automation/executions/${id}`, { method: "DELETE" });
}

export async function importAutomationExecution(
  projectId: string,
  input: {
    name: string;
    format: string;
    report: File;
    auto_create_defects?: boolean;
    map_test_cases?: boolean;
  },
): Promise<IngestResult> {
  const formData = new FormData();
  formData.append("project_id", projectId);
  formData.append("name", input.name);
  formData.append("format", input.format);
  formData.append("report", input.report);
  if (input.auto_create_defects) formData.append("auto_create_defects", "true");
  if (input.map_test_cases) formData.append("map_test_cases", "true");
  return apiFetch(`/api/v1/automation/projects/${projectId}/executions`, {
    method: "POST",
    body: formData,
  });
}

export async function listAutomationArtifacts(
  executionId: string,
  params?: { cursor?: string; limit?: number },
): Promise<{ data: AutomationArtifact[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams();
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/automation/executions/${executionId}/artifacts?${searchParams.toString()}`) as Promise<
    PaginatedResponse<AutomationArtifact>
  >;
}

export async function uploadAutomationArtifact(
  executionId: string,
  input: {
    workspace_id: string;
    file: File;
    kind?: string;
    test_run_item_id?: string;
  },
): Promise<AutomationArtifact> {
  const formData = new FormData();
  formData.append("workspace_id", input.workspace_id);
  formData.append("file", input.file);
  if (input.kind) formData.append("kind", input.kind);
  if (input.test_run_item_id) formData.append("test_run_item_id", input.test_run_item_id);
  return apiFetch(`/api/v1/automation/executions/${executionId}/artifacts`, {
    method: "POST",
    body: formData,
  });
}

export function getAutomationArtifactDownloadUrl(id: string): string {
  return `/api/v1/automation/artifacts/${id}`;
}

export async function listAutomationLogs(
  executionId: string,
  params?: { cursor?: string; limit?: number },
): Promise<{ data: AutomationLog[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams();
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/automation/executions/${executionId}/logs?${searchParams.toString()}`) as Promise<
    PaginatedResponse<AutomationLog>
  >;
}

export async function addAutomationLog(
  executionId: string,
  input: { workspace_id: string; level: string; message: string; logged_at?: string },
): Promise<AutomationLog> {
  const searchParams = new URLSearchParams({ workspace_id: input.workspace_id });
  return apiFetch(`/api/v1/automation/executions/${executionId}/logs?${searchParams.toString()}`, {
    method: "POST",
    body: JSON.stringify({
      level: input.level,
      message: input.message,
      logged_at: input.logged_at,
    }),
  });
}
