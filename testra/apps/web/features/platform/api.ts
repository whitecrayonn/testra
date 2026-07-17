import { apiFetch } from "@/lib/api";
import type { Organization, Workspace, Project, APIKey, User } from "@/types/platform";

export async function listOrganizations(): Promise<Organization[]> {
  return apiFetch("/api/v1/organizations");
}

export async function getOrganization(id: string): Promise<Organization> {
  return apiFetch(`/api/v1/organizations/${id}`);
}

export async function listWorkspaces(organizationId: string): Promise<Workspace[]> {
  return apiFetch(`/api/v1/workspaces?organization_id=${organizationId}`);
}

export async function getWorkspace(id: string): Promise<Workspace> {
  return apiFetch(`/api/v1/workspaces/${id}`);
}

export async function listProjects(workspaceId: string): Promise<Project[]> {
  return apiFetch(`/api/v1/projects?workspace_id=${workspaceId}`);
}

export async function createProject(input: {
  workspace_id: string;
  name: string;
  key: string;
  description?: string;
}): Promise<Project> {
  return apiFetch("/api/v1/projects", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function getProject(id: string): Promise<Project> {
  return apiFetch(`/api/v1/projects/${id}`);
}

export async function listAPIKeys(workspaceId: string): Promise<APIKey[]> {
  return apiFetch(`/api/v1/api-keys?workspace_id=${workspaceId}`);
}

export async function createAPIKey(input: {
  workspace_id: string;
  name: string;
  scopes?: string[];
  expires_in_days?: number;
}): Promise<{ api_key: APIKey; raw_key: string }> {
  return apiFetch("/api/v1/api-keys", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function revokeAPIKey(id: string): Promise<void> {
  return apiFetch(`/api/v1/api-keys/${id}`, {
    method: "DELETE",
  });
}

export async function getCurrentUser(): Promise<User> {
  return apiFetch("/api/v1/auth/me");
}
