import { apiFetch } from "@/lib/api";
import type {
  APICollection,
  APIFolder,
  APIEnvironment,
  APIRequest,
  APIRequestHistory,
  AuthConfig,
  ExecutionResponse,
  KeyValuePair,
  PaginationMeta,
} from "@/types/apitesting";

interface PaginatedResponse<T> {
  data: T[];
  meta: PaginationMeta;
}

export async function listCollections(
  workspaceId: string,
  params?: { cursor?: string; limit?: number },
): Promise<{ data: APICollection[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams({ workspace_id: workspaceId });
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/api-collections?${searchParams.toString()}`) as Promise<
    PaginatedResponse<APICollection>
  >;
}

export async function getCollection(id: string): Promise<APICollection> {
  return apiFetch(`/api/v1/api-collections/${id}`);
}

export async function createCollection(input: {
  workspace_id: string;
  name: string;
  description?: string;
}): Promise<APICollection> {
  return apiFetch("/api/v1/api-collections", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateCollection(
  id: string,
  input: { name: string; description?: string },
): Promise<APICollection> {
  return apiFetch(`/api/v1/api-collections/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function deleteCollection(id: string): Promise<void> {
  await apiFetch(`/api/v1/api-collections/${id}`, { method: "DELETE" });
}

export async function listFolders(
  collectionId: string,
  params?: { parentId?: string; cursor?: string; limit?: number },
): Promise<{ data: APIFolder[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams({ collection_id: collectionId });
  if (params?.parentId) searchParams.set("parent_id", params.parentId);
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/api-folders?${searchParams.toString()}`) as Promise<
    PaginatedResponse<APIFolder>
  >;
}

export async function createFolder(input: {
  workspace_id: string;
  collection_id: string;
  parent_id?: string;
  name: string;
}): Promise<APIFolder> {
  return apiFetch("/api/v1/api-folders", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateFolder(
  id: string,
  input: { parent_id?: string; name: string },
): Promise<APIFolder> {
  return apiFetch(`/api/v1/api-folders/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function deleteFolder(id: string): Promise<void> {
  await apiFetch(`/api/v1/api-folders/${id}`, { method: "DELETE" });
}

export async function listEnvironments(
  workspaceId: string,
  params?: { cursor?: string; limit?: number },
): Promise<{ data: APIEnvironment[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams({ workspace_id: workspaceId });
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/api-environments?${searchParams.toString()}`) as Promise<
    PaginatedResponse<APIEnvironment>
  >;
}

export async function createEnvironment(input: {
  workspace_id: string;
  name: string;
  variables?: KeyValuePair[];
}): Promise<APIEnvironment> {
  return apiFetch("/api/v1/api-environments", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateEnvironment(
  id: string,
  input: { name: string; variables?: KeyValuePair[] },
): Promise<APIEnvironment> {
  return apiFetch(`/api/v1/api-environments/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function deleteEnvironment(id: string): Promise<void> {
  await apiFetch(`/api/v1/api-environments/${id}`, { method: "DELETE" });
}

export async function listRequests(
  collectionId: string,
  params?: { folderId?: string; cursor?: string; limit?: number },
): Promise<{ data: APIRequest[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams({ collection_id: collectionId });
  if (params?.folderId) searchParams.set("folder_id", params.folderId);
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/api-requests?${searchParams.toString()}`) as Promise<
    PaginatedResponse<APIRequest>
  >;
}

export async function searchRequests(
  workspaceId: string,
  query: string,
  params?: { cursor?: string; limit?: number },
): Promise<{ data: APIRequest[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams({ workspace_id: workspaceId, q: query });
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/api-requests/search?${searchParams.toString()}`) as Promise<
    PaginatedResponse<APIRequest>
  >;
}

export async function getRequest(id: string): Promise<APIRequest> {
  return apiFetch(`/api/v1/api-requests/${id}`);
}

export async function createRequest(input: {
  workspace_id: string;
  collection_id: string;
  folder_id?: string;
  environment_id?: string;
  name: string;
  method: string;
  url: string;
  headers?: KeyValuePair[];
  query_params?: KeyValuePair[];
  auth_type?: string;
  auth_config?: AuthConfig;
  body_type?: string;
  body_content?: string;
  variables?: KeyValuePair[];
}): Promise<APIRequest> {
  return apiFetch("/api/v1/api-requests", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateRequest(
  id: string,
  input: {
    collection_id: string;
    folder_id?: string;
    environment_id?: string;
    name: string;
    method: string;
    url: string;
    headers?: KeyValuePair[];
    query_params?: KeyValuePair[];
    auth_type?: string;
    auth_config?: AuthConfig;
    body_type?: string;
    body_content?: string;
    variables?: KeyValuePair[];
  },
): Promise<APIRequest> {
  return apiFetch(`/api/v1/api-requests/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function deleteRequest(id: string): Promise<void> {
  await apiFetch(`/api/v1/api-requests/${id}`, { method: "DELETE" });
}

export async function executeRequest(input: {
  workspace_id: string;
  request_id?: string;
  environment_id?: string;
  request?: {
    name?: string;
    method: string;
    url: string;
    headers?: KeyValuePair[];
    query_params?: KeyValuePair[];
    auth_type?: string;
    auth_config?: AuthConfig;
    body_type?: string;
    body_content?: string;
    variables?: KeyValuePair[];
    environment_id?: string;
  };
  save?: boolean;
}): Promise<ExecutionResponse> {
  return apiFetch("/api/v1/api-executions", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function listRequestHistory(
  requestId: string,
  params?: { cursor?: string; limit?: number },
): Promise<{ data: APIRequestHistory[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams();
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/api-requests/${requestId}/history?${searchParams.toString()}`) as Promise<
    PaginatedResponse<APIRequestHistory>
  >;
}

export async function listExecutionHistory(
  workspaceId: string,
  params?: { requestId?: string; cursor?: string; limit?: number },
): Promise<{ data: APIRequestHistory[]; meta: PaginationMeta }> {
  const searchParams = new URLSearchParams({ workspace_id: workspaceId });
  if (params?.requestId) searchParams.set("request_id", params.requestId);
  if (params?.cursor) searchParams.set("cursor", params.cursor);
  if (params?.limit) searchParams.set("limit", String(params.limit));
  return apiFetch(`/api/v1/api-executions?${searchParams.toString()}`) as Promise<
    PaginatedResponse<APIRequestHistory>
  >;
}
