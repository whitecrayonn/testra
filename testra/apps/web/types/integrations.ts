export type IntegrationType = "jira" | "github" | "gitlab" | "slack" | "webhook";

export interface Integration {
  id: string;
  workspace_id: string;
  type: IntegrationType;
  name: string;
  config: Record<string, string>;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface IntegrationEvent {
  id: string;
  workspace_id: string;
  integration_id?: string;
  event_type: string;
  payload: Record<string, unknown>;
  status: string;
  external_id: string;
  created_at: string;
}
