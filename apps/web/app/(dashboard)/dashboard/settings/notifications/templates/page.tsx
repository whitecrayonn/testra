"use client";

import { useEffect, useState } from "react";
import { FileText, Plus, Trash2, Loader2, Save } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/ui/page-header";
import { listTemplates, createTemplate, updateTemplate, deleteTemplate } from "@/features/notifications/api";
import type { NotificationTemplate } from "@/types/notifications";

const channelTypes: NotificationTemplate["channel_type"][] = ["email", "slack", "teams", "webhook", "in_app"];

export default function NotificationTemplatesPage() {
  const [templates, setTemplates] = useState<NotificationTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const [form, setForm] = useState<{
    name: string;
    event_type: string;
    channel_type: NotificationTemplate["channel_type"];
    subject: string;
    body: string;
  }>({
    name: "",
    event_type: "",
    channel_type: "email",
    subject: "",
    body: "",
  });

  async function load() {
    try {
      const { data } = await listTemplates();
      setTemplates(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load templates");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!form.name.trim() || !form.event_type.trim()) return;
    setSaving("new");
    try {
      const { data } = await createTemplate(form);
      setTemplates((prev) => [data, ...prev]);
      setForm({ name: "", event_type: "", channel_type: "email", subject: "", body: "" });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create template");
    } finally {
      setSaving(null);
    }
  }

  async function handleUpdate(t: NotificationTemplate) {
    setSaving(t.id);
    try {
      const { data } = await updateTemplate(t.id, t);
      setTemplates((prev) => prev.map((x) => (x.id === data.id ? data : x)));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update template");
    } finally {
      setSaving(null);
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Delete this template?")) return;
    try {
      await deleteTemplate(id);
      setTemplates((prev) => prev.filter((x) => x.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete template");
    }
  }

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-slate-400" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Notification Templates"
        description="Manage templates for emails, Slack, Teams, webhooks, and in-app notifications."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "Notifications", href: "/dashboard/settings/notifications" },
          { label: "Templates" },
        ]}
      />

      {error && <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5 text-brand-600" />
            Templates
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <form onSubmit={handleCreate} className="space-y-4 rounded-lg border border-slate-200 p-4">
            <h3 className="font-medium text-slate-900">Create template</h3>
            <div className="grid gap-4 sm:grid-cols-3">
              <Input label="Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
              <Input label="Event type" value={form.event_type} onChange={(e) => setForm({ ...form, event_type: e.target.value })} required />
              <div>
                <label className="mb-1 block text-sm font-medium text-slate-700">Channel</label>
                <select
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                  value={form.channel_type}
                  onChange={(e) => setForm({ ...form, channel_type: e.target.value as NotificationTemplate["channel_type"] })}
                >
                  {channelTypes.map((c) => (
                    <option key={c} value={c}>{c}</option>
                  ))}
                </select>
              </div>
            </div>
            <Input label="Subject" value={form.subject} onChange={(e) => setForm({ ...form, subject: e.target.value })} />
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-700">Body</label>
              <textarea
                className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                rows={4}
                value={form.body}
                onChange={(e) => setForm({ ...form, body: e.target.value })}
              />
            </div>
            <Button type="submit" loading={saving === "new"}>
              <Plus className="mr-2 h-4 w-4" />
              Add template
            </Button>
          </form>

          <div className="space-y-4">
            {templates.map((t) => (
              <div key={t.id} className="space-y-3 rounded-lg border border-slate-200 p-4">
                <div className="grid gap-4 sm:grid-cols-3">
                  <Input label="Name" value={t.name} onChange={(e) => setTemplates((prev) => prev.map((x) => x.id === t.id ? { ...x, name: e.target.value } : x))} />
                  <Input label="Event type" value={t.event_type} onChange={(e) => setTemplates((prev) => prev.map((x) => x.id === t.id ? { ...x, event_type: e.target.value } : x))} />
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-700">Channel</label>
                    <select
                      className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                      value={t.channel_type}
                      onChange={(e) => setTemplates((prev) => prev.map((x) => x.id === t.id ? { ...x, channel_type: e.target.value as NotificationTemplate["channel_type"] } : x))}
                    >
                      {channelTypes.map((c) => (
                        <option key={c} value={c}>{c}</option>
                      ))}
                    </select>
                  </div>
                </div>
                <Input label="Subject" value={t.subject} onChange={(e) => setTemplates((prev) => prev.map((x) => x.id === t.id ? { ...x, subject: e.target.value } : x))} />
                <div>
                  <label className="mb-1 block text-sm font-medium text-slate-700">Body</label>
                  <textarea
                    className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                    rows={4}
                    value={t.body}
                    onChange={(e) => setTemplates((prev) => prev.map((x) => x.id === t.id ? { ...x, body: e.target.value } : x))}
                  />
                </div>
                <div className="flex gap-2">
                  <Button size="sm" loading={saving === t.id} onClick={() => handleUpdate(t)}>
                    <Save className="mr-2 h-4 w-4" />
                    Save
                  </Button>
                  <Button size="sm" variant="danger" onClick={() => handleDelete(t.id)}>
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
