"use client";

import { useEffect, useState } from "react";
import { Loader2, RefreshCw, Play, Archive } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { listIntegrationEvents, retryEvent, replayDeadLetterEvent } from "@/features/integrations/api";
import type { IntegrationEvent } from "@/types/integrations";

export default function RetryQueuePage() {
  const [events, setEvents] = useState<IntegrationEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [processing, setProcessing] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function load() {
    const workspaceId = localStorage.getItem("testra_workspace_id") || "";
    if (!workspaceId) {
      setLoading(false);
      setError("No workspace selected");
      return;
    }
    try {
      const [failed, dead] = await Promise.all([
        listIntegrationEvents(workspaceId, "failed"),
        listIntegrationEvents(workspaceId, "dead_letter"),
      ]);
      setEvents([...failed, ...dead]);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load retry queue");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, []);

  async function handleRetry(eventId: string) {
    setProcessing(eventId);
    try {
      await retryEvent(eventId);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Retry failed");
    } finally {
      setProcessing(null);
    }
  }

  async function handleReplay(eventId: string) {
    setProcessing(eventId);
    try {
      await replayDeadLetterEvent(eventId);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Replay failed");
    } finally {
      setProcessing(null);
    }
  }

  function statusVariant(status: string) {
    switch (status) {
      case "sent":
        return "success";
      case "failed":
        return "danger";
      case "dead_letter":
        return "warning";
      default:
        return "neutral";
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
        title="Retry Queue"
        description="Retry failed events or replay dead-letter events."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "Integrations", href: "/dashboard/settings/integrations" },
          { label: "Retry Queue" },
        ]}
      />

      {error && <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <RefreshCw className="h-5 w-5 text-brand-600" />
            Failed and Dead-letter Events
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <Button size="sm" variant="secondary" onClick={load}>
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>

          {events.length === 0 ? (
            <p className="text-sm text-slate-500">No events in retry queue.</p>
          ) : (
            <div className="space-y-2">
              {events.map((e) => (
                <div key={e.id} className="flex flex-col gap-2 rounded-lg border border-slate-200 p-3 text-sm sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <div className="flex flex-wrap items-center gap-2">
                      <Badge variant={statusVariant(e.status)}>{e.status}</Badge>
                      <span className="font-medium">{e.event_type}</span>
                      <span className="text-slate-500">{new Date(e.created_at).toLocaleString()}</span>
                    </div>
                    {e.retry_count > 0 && <p className="text-slate-500">Retries: {e.retry_count}</p>}
                    {!!e.payload?.error && <p className="text-red-600 break-all">{String(e.payload.error)}</p>}
                  </div>
                  <div className="flex items-center gap-2">
                    {e.status === "failed" && (
                      <Button size="sm" variant="secondary" loading={processing === e.id} onClick={() => handleRetry(e.id)}>
                        <Play className="mr-2 h-3 w-3" />
                        Retry
                      </Button>
                    )}
                    {e.status === "dead_letter" && (
                      <Button size="sm" variant="secondary" loading={processing === e.id} onClick={() => handleReplay(e.id)}>
                        <Archive className="mr-2 h-3 w-3" />
                        Replay
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
