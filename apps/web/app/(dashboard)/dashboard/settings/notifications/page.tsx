"use client";

import { useState, useEffect, useCallback } from "react";
import { Bell, Plus, Trash2, Mail, MessageSquare, AppWindow, Globe } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import {
  getPreferences,
  updatePreferences,
  listChannels,
  createChannel,
  deleteChannel,
} from "@/features/notifications/api";
import type { NotificationPreferences, NotificationChannel, NotificationChannelType } from "@/types/notifications";

const channelIcons: Record<NotificationChannelType, typeof Mail> = {
  email: Mail,
  slack: MessageSquare,
  teams: AppWindow,
  webhook: Globe,
};

export default function NotificationsSettingsPage() {
  const [preferences, setPreferences] = useState<NotificationPreferences | null>(null);
  const [channels, setChannels] = useState<NotificationChannel[]>([]);
  const [loadingPrefs, setLoadingPrefs] = useState(true);
  const [loadingChannels, setLoadingChannels] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [newType, setNewType] = useState<NotificationChannelType>("email");
  const [newName, setNewName] = useState("");
  const [newTo, setNewTo] = useState("");
  const [newUrl, setNewUrl] = useState("");
  const [creating, setCreating] = useState(false);

  const fetchData = useCallback(async () => {
    setLoadingPrefs(true);
    setLoadingChannels(true);
    setError(null);
    try {
      const [{ data: prefs }, { data: chs }] = await Promise.all([
        getPreferences(),
        listChannels(),
      ]);
      setPreferences(prefs);
      setChannels(chs);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load notification settings");
    } finally {
      setLoadingPrefs(false);
      setLoadingChannels(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  async function handleSavePreferences() {
    if (!preferences) return;
    setSaving(true);
    setError(null);
    try {
      const { data } = await updatePreferences({
        in_app_enabled: preferences.in_app_enabled,
        email_enabled: preferences.email_enabled,
        slack_enabled: preferences.slack_enabled,
        teams_enabled: preferences.teams_enabled,
        webhook_enabled: preferences.webhook_enabled,
      });
      setPreferences(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save preferences");
    } finally {
      setSaving(false);
    }
  }

  async function handleCreateChannel(e: React.FormEvent) {
    e.preventDefault();
    const workspaceId =
      typeof window !== "undefined"
        ? localStorage.getItem("testra_workspace_id") || ""
        : "";
    if (!workspaceId || !newName.trim()) return;

    const config: Record<string, string> =
      newType === "email" ? { to: newTo } : { url: newUrl };

    setCreating(true);
    setError(null);
    try {
      const { data } = await createChannel({
        workspace_id: workspaceId,
        type: newType,
        name: newName.trim(),
        config,
      });
      setChannels((prev) => [...prev, data]);
      setNewName("");
      setNewTo("");
      setNewUrl("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create channel");
    } finally {
      setCreating(false);
    }
  }

  async function handleDeleteChannel(id: string) {
    if (!confirm("Delete this notification channel?")) return;
    try {
      await deleteChannel(id);
      setChannels((prev) => prev.filter((c) => c.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete channel");
    }
  }

  function toggle<K extends keyof NotificationPreferences>(key: K) {
    setPreferences((prev) => (prev ? { ...prev, [key]: !prev[key] } : prev));
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Notifications"
        description="Choose which events trigger email, in-app, and webhook notifications."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "Notifications" },
        ]}
      />

      {error && (
        <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Notification preferences</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {loadingPrefs ? (
            <Skeleton className="h-8 w-full" count={3} />
          ) : preferences ? (
            <>
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium text-slate-900">In-app</p>
                  <p className="text-sm text-slate-500">Show notifications in the application.</p>
                </div>
                <Switch
                  checked={preferences.in_app_enabled}
                  onCheckedChange={() => toggle("in_app_enabled")}
                />
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium text-slate-900">Email</p>
                  <p className="text-sm text-slate-500">Send notifications via configured email channels.</p>
                </div>
                <Switch
                  checked={preferences.email_enabled}
                  onCheckedChange={() => toggle("email_enabled")}
                />
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium text-slate-900">Slack</p>
                  <p className="text-sm text-slate-500">Forward notifications to Slack webhooks.</p>
                </div>
                <Switch
                  checked={preferences.slack_enabled}
                  onCheckedChange={() => toggle("slack_enabled")}
                />
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium text-slate-900">Microsoft Teams</p>
                  <p className="text-sm text-slate-500">Forward notifications to Teams connectors.</p>
                </div>
                <Switch
                  checked={preferences.teams_enabled}
                  onCheckedChange={() => toggle("teams_enabled")}
                />
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium text-slate-900">Webhook</p>
                  <p className="text-sm text-slate-500">POST notification payloads to custom URLs.</p>
                </div>
                <Switch
                  checked={preferences.webhook_enabled}
                  onCheckedChange={() => toggle("webhook_enabled")}
                />
              </div>
              <Button onClick={handleSavePreferences} loading={saving}>
                Save preferences
              </Button>
            </>
          ) : (
            <p className="text-sm text-slate-500">No preferences available.</p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Notification channels</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {loadingChannels ? (
            <Skeleton className="h-20 w-full" count={3} />
          ) : channels.length === 0 ? (
            <EmptyState
              icon={Bell}
              title="No channels"
              description="Add email, Slack, Teams or webhook channels to dispatch notifications."
            />
          ) : (
            <div className="space-y-3">
              {channels.map((channel) => {
                const Icon = channelIcons[channel.type] || Globe;
                return (
                  <Card key={channel.id} className="p-4">
                    <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                      <div className="flex items-start gap-3">
                        <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-brand-50 text-brand-600">
                          <Icon className="h-5 w-5" />
                        </div>
                        <div>
                          <h3 className="font-semibold text-slate-900">{channel.name}</h3>
                          <Badge variant="neutral" className="mt-1 capitalize">
                            {channel.type}
                          </Badge>
                          <p className="mt-1 text-xs text-slate-500 break-all">
                            {channel.type === "email" ? channel.config.to : channel.config.url}
                          </p>
                        </div>
                      </div>
                      <Button
                        variant="danger"
                        size="sm"
                        onClick={() => handleDeleteChannel(channel.id)}
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        Delete
                      </Button>
                    </div>
                  </Card>
                );
              })}
            </div>
          )}

          <form onSubmit={handleCreateChannel} className="space-y-4 rounded-lg border border-slate-200 p-4">
            <h3 className="font-medium text-slate-900">Add channel</h3>
            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <label className="mb-1 block text-sm font-medium text-slate-700">Channel type</label>
                <select
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                  value={newType}
                  onChange={(e) => setNewType(e.target.value as NotificationChannelType)}
                >
                  <option value="email">Email</option>
                  <option value="slack">Slack</option>
                  <option value="teams">Microsoft Teams</option>
                  <option value="webhook">Webhook</option>
                </select>
              </div>
              <Input
                label="Name"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="e.g. Engineering Slack"
                required
              />
            </div>
            {newType === "email" ? (
              <Input
                label="To"
                type="email"
                value={newTo}
                onChange={(e) => setNewTo(e.target.value)}
                placeholder="team@example.com"
                required
              />
            ) : (
              <Input
                label="Webhook URL"
                type="url"
                value={newUrl}
                onChange={(e) => setNewUrl(e.target.value)}
                placeholder="https://hooks.slack.com/services/..."
                required
              />
            )}
            <Button type="submit" loading={creating}>
              <Plus className="mr-2 h-4 w-4" />
              Add channel
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
