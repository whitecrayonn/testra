import { apiFetch } from "@/lib/api";
import type { Integration, IntegrationEvent, IntegrationType } from "@/types/integrations";

export async function listIntegrations(workspaceId: string): Promise<Integration[]> {
  return apiFetch(`/api/v1/integrations?workspace_id=${workspaceId}`);
}

export async function getIntegration(id: string): Promise<Integration> {
  return apiFetch(`/api/v1/integrations/${id}`);
}

export async function createIntegration(input: {
  workspace_id: string;
  type: IntegrationType;
  name: string;
  config: Record<string, string>;
  enabled: boolean;
}): Promise<Integration> {
  return apiFetch("/api/v1/integrations", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateIntegration(
  id: string,
  input: Partial<Omit<Integration, "id" | "workspace_id" | "created_at" | "updated_at">>,
): Promise<Integration> {
  return apiFetch(`/api/v1/integrations/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function deleteIntegration(id: string): Promise<void> {
  await apiFetch(`/api/v1/integrations/${id}`, { method: "DELETE" });
}

export async function testIntegration(id: string): Promise<void> {
  await apiFetch(`/api/v1/integrations/${id}/test`, { method: "POST" });
}

export async function listIntegrationEvents(workspaceId: string): Promise<IntegrationEvent[]> {
  return apiFetch(`/api/v1/integration-events?workspace_id=${workspaceId}`);
}
