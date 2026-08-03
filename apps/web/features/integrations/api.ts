import { apiFetch } from "@/lib/api";
import type { Integration, IntegrationEvent, IntegrationHealth, IntegrationType } from "@/types/integrations";

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

export async function listIntegrationEvents(workspaceId: string, status?: string): Promise<IntegrationEvent[]> {
  let url = `/api/v1/integration-events?workspace_id=${workspaceId}`;
  if (status) url += `&status=${status}`;
  return apiFetch(url);
}

export async function getIntegrationHealth(id: string): Promise<IntegrationHealth> {
  return apiFetch(`/api/v1/integrations/${id}/health`);
}

export async function setIntegrationEnabled(id: string, enabled: boolean): Promise<Integration> {
  return apiFetch(`/api/v1/integrations/${id}/${enabled ? "enable" : "disable"}`, { method: "POST" });
}

export async function retryEvent(id: string): Promise<IntegrationEvent> {
  return apiFetch(`/api/v1/integration-events/${id}/retry`, { method: "POST" });
}

export async function replayDeadLetterEvent(id: string): Promise<IntegrationEvent> {
  return apiFetch(`/api/v1/integration-events/${id}/replay`, { method: "POST" });
}

export async function dispatchIntegrationEvent(
  integrationId: string,
  workspaceId: string,
  eventType: string,
  payload: Record<string, unknown>
): Promise<IntegrationEvent> {
  return apiFetch("/api/v1/integration-events/dispatch", {
    method: "POST",
    body: JSON.stringify({ integration_id: integrationId, workspace_id: workspaceId, event_type: eventType, payload }),
  });
}
