"use client";

import { useEffect, useState } from "react";
import { Plug, Plus, Trash2, Loader2, TestTube } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import Link from "next/link";
import {
  listIntegrations,
  createIntegration,
  deleteIntegration,
  testIntegration,
  setIntegrationEnabled,
} from "@/features/integrations/api";
import type { Integration, IntegrationType } from "@/types/integrations";

const integrationTypes: IntegrationType[] = [
  "jira",
  "github",
  "gitlab",
  "bitbucket",
  "azure_devops",
  "linear",
  "slack",
  "discord",
  "webhook",
  "smtp",
];

export default function IntegrationsSettingsPage() {
  const [integrations, setIntegrations] = useState<Integration[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [testing, setTesting] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const [type, setType] = useState<IntegrationType>("webhook");
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [token, setToken] = useState("");

  async function load() {
    const workspaceId = localStorage.getItem("testra_workspace_id") || "";
    if (!workspaceId) {
      setLoading(false);
      setError("No workspace selected");
      return;
    }
    try {
      const data = await listIntegrations(workspaceId);
      setIntegrations(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load integrations");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    const workspaceId = localStorage.getItem("testra_workspace_id") || "";
    if (!workspaceId || !name.trim()) return;
    setCreating(true);
    setError(null);
    const config: Record<string, string> = { url };
    if (token) config.token = token;
    try {
      const integration = await createIntegration({
        workspace_id: workspaceId,
        type,
        name: name.trim(),
        config,
        enabled: true,
      });
      setIntegrations((prev) => [...prev, integration]);
      setName("");
      setUrl("");
      setToken("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create integration");
    } finally {
      setCreating(false);
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Delete this integration?")) return;
    try {
      await deleteIntegration(id);
      setIntegrations((prev) => prev.filter((i) => i.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete integration");
    }
  }

  async function handleTest(id: string) {
    setTesting(id);
    try {
      await testIntegration(id);
      alert("Integration test succeeded");
      load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Integration test failed");
    } finally {
      setTesting(null);
    }
  }

  async function handleToggleEnabled(integration: Integration) {
    try {
      const updated = await setIntegrationEnabled(integration.id, !integration.enabled);
      setIntegrations((prev) => prev.map((i) => (i.id === integration.id ? updated : i)));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update integration");
    }
  }

  function healthVariant(status?: string): "success" | "danger" | "neutral" {
    switch (status) {
      case "healthy":
        return "success";
      case "unhealthy":
        return "danger";
      default:
        return "neutral";
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Integrations"
        description="Connect Jira, GitHub, GitLab, Slack, or webhooks to Testra."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "Integrations" },
        ]}
      />

      {error && <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Plug className="h-5 w-5 text-brand-600" />
            Connected integrations
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {loading ? (
            <div className="flex h-24 items-center justify-center">
              <Loader2 className="h-6 w-6 animate-spin text-slate-400" />
            </div>
          ) : integrations.length === 0 ? (
            <EmptyState icon={Plug} title="No integrations" description="Add your first integration below." />
          ) : (
            <div className="space-y-3">
              {integrations.map((integration) => (
                <div key={integration.id} className="flex flex-col gap-3 rounded-lg border border-slate-200 p-4 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="font-semibold text-slate-900">{integration.name}</h3>
                      {!integration.enabled && <Badge variant="neutral">Disabled</Badge>}
                    </div>
                    <div className="mt-1 flex flex-wrap items-center gap-2">
                      <Badge variant="neutral" className="capitalize">{integration.type}</Badge>
                      {integration.health_status && <Badge variant={healthVariant(integration.health_status) as "success" | "danger" | "neutral"}>{integration.health_status}</Badge>}
                    </div>
                    <p className="mt-1 text-xs text-slate-500 break-all">{integration.config.url}</p>
                    {integration.last_error && <p className="text-xs text-red-600">{integration.last_error}</p>}
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => handleToggleEnabled(integration)}
                    >
                      {integration.enabled ? "Disable" : "Enable"}
                    </Button>
                    <Button
                      size="sm"
                      variant="secondary"
                      loading={testing === integration.id}
                      onClick={() => handleTest(integration.id)}
                    >
                      <TestTube className="mr-2 h-4 w-4" />
                      Test
                    </Button>
                    <Link href={`/dashboard/settings/integrations/${integration.id}`}>
                      <Button size="sm" variant="secondary">Details</Button>
                    </Link>
                    <Link href={`/dashboard/settings/integrations/${integration.id}/events`}>
                      <Button size="sm" variant="secondary">Logs</Button>
                    </Link>
                    <Button size="sm" variant="danger" onClick={() => handleDelete(integration.id)}>
                      <Trash2 className="mr-2 h-4 w-4" />
                      Delete
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}

          <form onSubmit={handleCreate} className="space-y-4 rounded-lg border border-slate-200 p-4">
            <h3 className="font-medium text-slate-900">Add integration</h3>
            <div className="grid gap-4 sm:grid-cols-3">
              <div>
                <label className="mb-1 block text-sm font-medium text-slate-700">Type</label>
                <select
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                  value={type}
                  onChange={(e) => setType(e.target.value as IntegrationType)}
                >
                  {integrationTypes.map((t) => (
                    <option key={t} value={t}>
                      {t}
                    </option>
                  ))}
                </select>
              </div>
              <Input label="Name" value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. Slack alerts" required />
              <Input label="URL" value={url} onChange={(e) => setUrl(e.target.value)} placeholder="https://..." required />
            </div>
            <Input label="Token / Secret (optional)" value={token} onChange={(e) => setToken(e.target.value)} placeholder="Auth token or webhook secret" />
            <Button type="submit" loading={creating}>
              <Plus className="mr-2 h-4 w-4" />
              Add integration
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
