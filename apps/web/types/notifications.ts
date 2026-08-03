export type NotificationType =
  | "system"
  | "test_run_completed"
  | "defect_assigned"
  | "mention";

export interface Notification {
  id: string;
  organization_id: string;
  user_id: string;
  type: NotificationType;
  title: string;
  body: string;
  link: string;
  read: boolean;
  created_at: string;
  updated_at: string;
}

export interface NotificationPreferences {
  organization_id: string;
  user_id: string;
  in_app_enabled: boolean;
  email_enabled: boolean;
  slack_enabled: boolean;
  teams_enabled: boolean;
  webhook_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export type NotificationChannelType = "email" | "slack" | "teams" | "webhook";

export interface NotificationChannel {
  id: string;
  organization_id: string;
  workspace_id: string;
  type: NotificationChannelType;
  name: string;
  config: Record<string, string>;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface NotificationTemplate {
  id: string;
  organization_id: string;
  name: string;
  event_type: string;
  channel_type: "email" | "slack" | "teams" | "webhook" | "in_app";
  subject: string;
  body: string;
  created_at: string;
  updated_at: string;
}

export interface NotificationHistory {
  id: string;
  organization_id: string;
  notification_id: string;
  channel_id?: string;
  channel_type: string;
  status: "pending" | "sent" | "failed";
  error_message: string;
  retry_count: number;
  created_at: string;
  updated_at: string;
}
