import { apiFetch } from "@/lib/api";
import type {
  Notification,
  NotificationChannel,
  NotificationHistory,
  NotificationPreferences,
  NotificationTemplate,
} from "@/types/notifications";

export interface PaginatedNotifications {
  data: Notification[];
  meta: {
    has_more: boolean;
    next_cursor?: string;
  };
}

function getWorkspaceId(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("testra_workspace_id");
}

export async function listNotifications(
  cursor?: string,
  read?: boolean,
  limit = 20,
): Promise<PaginatedNotifications> {
  const workspaceId = getWorkspaceId();
  if (!workspaceId) throw new Error("No workspace selected");

  const params = new URLSearchParams({ workspace_id: workspaceId, limit: String(limit) });
  if (cursor) params.set("cursor", cursor);
  if (read !== undefined) params.set("read", String(read));

  return apiFetch<PaginatedNotifications>(`/api/v1/notifications?${params.toString()}`);
}

export async function getUnreadCount(): Promise<{ unread_count: number }> {
  const workspaceId = getWorkspaceId();
  if (!workspaceId) return { unread_count: 0 };

  return apiFetch<{ unread_count: number }>(
    `/api/v1/notifications/unread-count?workspace_id=${workspaceId}`,
  );
}

export async function markRead(id: string, read: boolean): Promise<void> {
  await apiFetch(`/api/v1/notifications/${id}?workspace_id=${getWorkspaceId() ?? ""}`, {
    method: "PATCH",
    body: JSON.stringify({ read }),
  });
}

export async function deleteNotification(id: string): Promise<void> {
  await apiFetch(`/api/v1/notifications/${id}?workspace_id=${getWorkspaceId() ?? ""}`, {
    method: "DELETE",
  });
}

export async function getPreferences(): Promise<{ data: NotificationPreferences }> {
  const workspaceId = getWorkspaceId();
  if (!workspaceId) throw new Error("No workspace selected");

  return apiFetch<{ data: NotificationPreferences }>(
    `/api/v1/notification-preferences?workspace_id=${workspaceId}`,
  );
}

export async function updatePreferences(
  preferences: Omit<NotificationPreferences, "organization_id" | "user_id" | "created_at" | "updated_at">,
): Promise<{ data: NotificationPreferences }> {
  const workspaceId = getWorkspaceId();
  if (!workspaceId) throw new Error("No workspace selected");

  return apiFetch<{ data: NotificationPreferences }>(
    `/api/v1/notification-preferences?workspace_id=${workspaceId}`,
    {
      method: "PUT",
      body: JSON.stringify(preferences),
    },
  );
}

export async function listChannels(): Promise<{ data: NotificationChannel[] }> {
  const workspaceId = getWorkspaceId();
  if (!workspaceId) throw new Error("No workspace selected");

  return apiFetch<{ data: NotificationChannel[] }>(
    `/api/v1/notification-channels?workspace_id=${workspaceId}`,
  );
}

export async function createChannel(
  channel: Omit<NotificationChannel, "id" | "organization_id" | "created_by" | "created_at" | "updated_at">,
): Promise<{ data: NotificationChannel }> {
  return apiFetch<{ data: NotificationChannel }>("/api/v1/notification-channels", {
    method: "POST",
    body: JSON.stringify(channel),
  });
}

export async function updateChannel(
  id: string,
  channel: Partial<NotificationChannel>,
): Promise<{ data: NotificationChannel }> {
  return apiFetch<{ data: NotificationChannel }>(`/api/v1/notification-channels/${id}`, {
    method: "PUT",
    body: JSON.stringify(channel),
  });
}

export async function deleteChannel(id: string): Promise<void> {
  await apiFetch(`/api/v1/notification-channels/${id}`, { method: "DELETE" });
}

function getOrganizationId(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("testra_organization_id");
}

export async function listTemplates(
  eventType?: string,
  channelType?: string,
): Promise<{ data: NotificationTemplate[] }> {
  const params = new URLSearchParams({ organization_id: getOrganizationId() ?? "" });
  if (eventType) params.set("event_type", eventType);
  if (channelType) params.set("channel_type", channelType);
  return apiFetch<{ data: NotificationTemplate[] }>(`/api/v1/notification-templates?${params.toString()}`);
}

export async function createTemplate(
  template: Omit<NotificationTemplate, "id" | "organization_id" | "created_at" | "updated_at">,
): Promise<{ data: NotificationTemplate }> {
  const payload = { ...template, organization_id: getOrganizationId() ?? "" };
  return apiFetch<{ data: NotificationTemplate }>("/api/v1/notification-templates", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateTemplate(
  id: string,
  template: Partial<NotificationTemplate>,
): Promise<{ data: NotificationTemplate }> {
  return apiFetch<{ data: NotificationTemplate }>(`/api/v1/notification-templates/${id}`, {
    method: "PUT",
    body: JSON.stringify(template),
  });
}

export async function deleteTemplate(id: string): Promise<void> {
  await apiFetch(`/api/v1/notification-templates/${id}`, { method: "DELETE" });
}

export async function listNotificationHistory(notificationId: string): Promise<{ data: NotificationHistory[] }> {
  return apiFetch<{ data: NotificationHistory[] }>(`/api/v1/notification-history?notification_id=${notificationId}`);
}
