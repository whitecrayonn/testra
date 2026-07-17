import { apiFetch } from "@/lib/api";
import type {
  TestCase,
  TestCaseVersion,
  TestFolder,
  TestSuite,
  PaginationMeta,
} from "@/types/testmanagement";

interface PaginatedResponse<T> {
  data: T[];
  meta: PaginationMeta;
}

export async function listTestCases(
  projectId: string,
  params?: { suiteId?: string; cursor?: string; limit?: number },
): Promise<{ data: TestCase[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams({ project_id: projectId });
  if (params?.suiteId) searchParams.set("suite_id", params.suiteId);
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/test-cases?${searchParams.toString()}`) as Promise<
    PaginatedResponse<TestCase>
  >;
}

export async function searchTestCases(
  workspaceId: string,
  query: string,
  params?: { cursor?: string; limit?: number },
): Promise<{ data: TestCase[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams({
    workspace_id: workspaceId,
    q: query,
  });
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(
    `/api/v1/test-cases/search?${searchParams.toString()}`,
  ) as Promise<PaginatedResponse<TestCase>>;
}

export async function getTestCase(id: string): Promise<TestCase> {
  return apiFetch(`/api/v1/test-cases/${id}`);
}

export async function createTestCase(input: {
  workspace_id: string;
  project_id: string;
  suite_id?: string;
  title: string;
  description?: string;
  preconditions?: string;
  steps?: Array<{ action: string; expected: string; test_data?: string }>;
  status?: string;
  priority?: string;
  tags?: string[];
}): Promise<TestCase> {
  return apiFetch("/api/v1/test-cases", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateTestCase(
  id: string,
  input: {
    suite_id?: string;
    title: string;
    description?: string;
    preconditions?: string;
    steps?: Array<{ action: string; expected: string; test_data?: string }>;
    status?: string;
    priority?: string;
    tags?: string[];
  },
): Promise<TestCase> {
  return apiFetch(`/api/v1/test-cases/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function deleteTestCase(id: string): Promise<void> {
  await apiFetch(`/api/v1/test-cases/${id}`, { method: "DELETE" });
}

export async function listTestCaseVersions(
  id: string,
): Promise<TestCaseVersion[]> {
  return apiFetch(`/api/v1/test-cases/${id}/versions`);
}

export async function listTestFolders(
  workspaceId: string,
  parentId?: string,
): Promise<TestFolder[]> {
  const searchParams = new URLSearchParams({ workspace_id: workspaceId });
  if (parentId) searchParams.set("parent_id", parentId);
  return apiFetch(`/api/v1/test-folders?${searchParams.toString()}`);
}

export async function listTestSuites(
  workspaceId: string,
  folderId?: string,
): Promise<TestSuite[]> {
  const searchParams = new URLSearchParams({ workspace_id: workspaceId });
  if (folderId) searchParams.set("folder_id", folderId);
  return apiFetch(`/api/v1/test-suites?${searchParams.toString()}`);
}
