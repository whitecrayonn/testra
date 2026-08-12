"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { Plug, Loader2, TestTube, RefreshCw } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import {
  getIntegration,
  updateIntegration,
  testIntegration,
  setIntegrationEnabled,
  getIntegrationHealth,
} from "@/features/integrations/api";
import type { Integration, IntegrationHealth } from "@/types/integrations";

export default function IntegrationDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const [integration, setIntegration] = useState<Integration | null>(null);
  const [health, setHealth] = useState<IntegrationHealth | null>(null);
  const [loading, setLoading] = useState(true);
  const [testing, setTesting] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [name, setName] = useState("");
  const [configJson, setConfigJson] = useState("{}");

  const load = useCallback(async () => {
    try {
      const [i, h] = await Promise.all([getIntegration(id), getIntegrationHealth(id)]);
      setIntegration(i);
      setHealth(h);
      setName(i.name);
      setConfigJson(JSON.stringify(i.config, null, 2));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load integration");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleTest() {
    setTesting(true);
    try {
      await testIntegration(id);
      const h = await getIntegrationHealth(id);
      setHealth(h);
      alert("Integration test succeeded");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Integration test failed");
    } finally {
      setTesting(false);
    }
  }

  async function handleToggleEnabled() {
    if (!integration) return;
    try {
      const updated = await setIntegrationEnabled(id, !integration.enabled);
      setIntegration(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update integration");
    }
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    if (!integration) return;
    setSaving(true);
    try {
      const config = JSON.parse(configJson);
      const updated = await updateIntegration(id, { name: name.trim(), config });
      setIntegration(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save integration");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-slate-400" />
      </div>
    );
  }

  if (!integration) {
    return <div className="p-6 text-red-600">{error || "Integration not found"}</div>;
  }

  const healthVariant = health?.health_status === "healthy" ? "success" : health?.health_status === "unhealthy" ? "danger" : "neutral";

  return (
    <div className="space-y-6">
      <PageHeader
        title={integration.name}
        description="Manage connection settings, health, and status."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "Integrations", href: "/dashboard/settings/integrations" },
          { label: integration.name },
        ]}
      />

      {error && <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Plug className="h-5 w-5 text-brand-600" />
            Connection Details
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant="neutral" className="capitalize">{integration.type}</Badge>
            {integration.enabled ? <Badge variant="success">Enabled</Badge> : <Badge variant="neutral">Disabled</Badge>}
            {health && <Badge variant={healthVariant}>{health.health_status}</Badge>}
          </div>
          {integration.last_tested_at && (
            <p className="text-sm text-slate-500">Last tested: {new Date(integration.last_tested_at).toLocaleString()}</p>
          )}
          {integration.last_error && <p className="text-sm text-red-600">{integration.last_error}</p>}

          <form onSubmit={handleSave} className="space-y-4">
            <Input label="Name" value={name} onChange={(e) => setName(e.target.value)} required />
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-700">Config (JSON)</label>
              <textarea
                className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm font-mono"
                rows={8}
                value={configJson}
                onChange={(e) => setConfigJson(e.target.value)}
              />
            </div>
            <div className="flex flex-wrap gap-2">
              <Button type="submit" loading={saving}>
                <RefreshCw className="mr-2 h-4 w-4" />
                Save
              </Button>
              <Button type="button" variant="secondary" loading={testing} onClick={handleTest}>
                <TestTube className="mr-2 h-4 w-4" />
                Test Connection
              </Button>
              <Button type="button" variant="secondary" onClick={handleToggleEnabled}>
                {integration.enabled ? "Disable" : "Enable"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
