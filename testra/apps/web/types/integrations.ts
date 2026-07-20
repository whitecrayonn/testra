export type IntegrationType =
  | "jira"
  | "github"
  | "gitlab"
  | "bitbucket"
  | "azure_devops"
  | "linear"
  | "slack"
  | "discord"
  | "webhook"
  | "smtp";

export interface Integration {
  id: string;
  workspace_id: string;
  type: IntegrationType;
  name: string;
  config: Record<string, string>;
  enabled: boolean;
  health_status?: "healthy" | "unhealthy" | "unknown";
  last_tested_at?: string;
  last_error?: string;
  sync_status?: string;
  retry_count?: number;
  created_at: string;
  updated_at: string;
}

export interface IntegrationHealth {
  id: string;
  type: IntegrationType;
  name: string;
  health_status: string;
  last_tested_at?: string;
  last_error?: string;
}

export interface IntegrationEvent {
  id: string;
  workspace_id: string;
  integration_id?: string;
  event_type: string;
  payload: Record<string, unknown>;
  status: "pending" | "sent" | "failed" | "dead_letter" | "received";
  external_id: string;
  retry_count: number;
  created_at: string;
  updated_at: string;
}
