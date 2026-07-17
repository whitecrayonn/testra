"use client";

import { useState, useEffect, useCallback } from "react";
import { Key, Copy, Trash2, Check } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import { listAPIKeys, createAPIKey, revokeAPIKey } from "@/features/platform/api";
import type { APIKey } from "@/types/platform";

export default function ApiKeysSettingsPage() {
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [name, setName] = useState("");
  const [plaintext, setPlaintext] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const workspaceId =
    typeof window !== "undefined"
      ? localStorage.getItem("testra_workspace_id") || ""
      : "";

  const fetchKeys = useCallback(async () => {
    if (!workspaceId) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const data = await listAPIKeys(workspaceId);
      setKeys(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load API keys");
    } finally {
      setLoading(false);
    }
  }, [workspaceId]);

  useEffect(() => {
    fetchKeys();
  }, [fetchKeys]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!workspaceId || !name.trim()) return;
    setCreating(true);
    setError(null);
    setPlaintext(null);
    try {
      const result = await createAPIKey({ workspace_id: workspaceId, name: name.trim() });
      setKeys((prev) => [...prev, result.api_key]);
      setPlaintext(result.raw_key);
      setName("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create API key");
    } finally {
      setCreating(false);
    }
  }

  async function handleRevoke(id: string) {
    if (!confirm("Are you sure you want to revoke this API key? This cannot be undone.")) return;
    try {
      await revokeAPIKey(id);
      setKeys((prev) => prev.filter((k) => k.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to revoke API key");
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="API Keys"
        description="Create and manage scoped API keys for CI/CD integration."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "API Keys" },
        ]}
      />

      {error && (
        <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>
      )}

      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-brand-50 text-brand-600">
              <Key className="h-6 w-6" aria-hidden="true" />
            </div>
            <CardTitle>Create API key</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-end">
              <div className="flex-1">
                <Input
                  label="Key name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g. CI/CD deployment"
                  disabled={!workspaceId || creating}
                />
              </div>
              <Button type="submit" loading={creating} disabled={!workspaceId}>
                Generate key
              </Button>
            </div>
            {!workspaceId && (
              <p className="text-sm text-slate-500">Select a workspace before creating API keys.</p>
            )}
          </form>
        </CardContent>
      </Card>

      {plaintext && (
        <Card className="border-brand-200 bg-brand-50">
          <CardContent className="space-y-3 p-6">
            <p className="text-sm font-medium text-slate-900">Copy this key now. It will not be shown again.</p>
            <div className="flex items-center gap-2 rounded-lg border border-slate-300 bg-white p-3">
              <code className="flex-1 break-all text-sm font-mono text-slate-700">{plaintext}</code>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => copyToClipboard(plaintext)}
                aria-label="Copy API key"
              >
                {copied ? <Check className="h-4 w-4 text-green-600" /> : <Copy className="h-4 w-4" />}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {loading ? (
        <Skeleton className="h-20 w-full" count={3} />
      ) : keys.length === 0 ? (
        <EmptyState
          icon={Key}
          title="No API keys"
          description="Generate an API key to start sending results from CI/CD."
          action={{ label: "Create API key", onClick: () => document.getElementById("api-key-name")?.focus() }}
        />
      ) : (
        <div className="space-y-3">
          {keys.map((key) => (
            <Card key={key.id} className="p-4">
              <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <h3 className="font-semibold text-slate-900">{key.name}</h3>
                  <p className="text-xs text-slate-500">{key.prefix}••••••••</p>
                  <div className="mt-2 flex flex-wrap gap-2">
                    {key.scopes.map((scope) => (
                      <Badge key={scope} variant="neutral">
                        {scope}
                      </Badge>
                    ))}
                    {key.expires_at ? (
                      <Badge variant="warning">Expires {new Date(key.expires_at).toLocaleDateString()}</Badge>
                    ) : (
                      <Badge variant="neutral">No expiry</Badge>
                    )}
                  </div>
                </div>
                <Button
                  variant="danger"
                  size="sm"
                  onClick={() => handleRevoke(key.id)}
                  aria-label={`Revoke ${key.name}`}
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Revoke
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
