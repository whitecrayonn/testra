"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { Loader2, RefreshCw, Play, Archive, AlertCircle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { listIntegrationEvents, retryEvent, replayDeadLetterEvent, getIntegration } from "@/features/integrations/api";
import type { IntegrationEvent, Integration } from "@/types/integrations";

export default function IntegrationEventsPage() {
  const params = useParams();
  const id = params.id as string;

  const [integration, setIntegration] = useState<Integration | null>(null);
  const [events, setEvents] = useState<IntegrationEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<string>("");
  const [error, setError] = useState<string | null>(null);
  const [processing, setProcessing] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const [i, evts] = await Promise.all([getIntegration(id), listIntegrationEvents(id)]);
      setIntegration(i);
      setEvents(evts);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load events");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    load();
  }, [load]);

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

  const filtered = filter ? events.filter((e) => e.status === filter) : events;

  function statusVariant(status: string) {
    switch (status) {
      case "sent":
        return "success";
      case "failed":
        return "danger";
      case "dead_letter":
        return "warning";
      case "received":
        return "info";
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
        title={`${integration?.name || "Integration"} Events`}
        description="Webhook logs, retries, and dead-letter replay."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "Integrations", href: "/dashboard/settings/integrations" },
          { label: integration?.name || "...", href: `/dashboard/settings/integrations/${id}` },
          { label: "Events" },
        ]}
      />

      {error && <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Archive className="h-5 w-5 text-brand-600" />
            Event Log
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap gap-2">
            {["", "pending", "sent", "failed", "dead_letter", "received"].map((s) => (
              <Button key={s} size="sm" variant={filter === s ? "primary" : "secondary"} onClick={() => setFilter(s)}>
                {s === "" ? "All" : s.replace("_", " ")}
              </Button>
            ))}
            <Button size="sm" variant="ghost" onClick={load}>
              <RefreshCw className="mr-2 h-4 w-4" />
              Refresh
            </Button>
          </div>

          {filtered.length === 0 ? (
            <p className="text-sm text-slate-500">No events found.</p>
          ) : (
            <div className="space-y-2">
              {filtered.map((e) => (
                <div key={e.id} className="rounded-lg border border-slate-200 p-3 text-sm">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <Badge variant={statusVariant(e.status)}>{e.status}</Badge>
                      <span className="font-medium">{e.event_type}</span>
                      <span className="text-slate-500">{new Date(e.created_at).toLocaleString()}</span>
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
                          <RefreshCw className="mr-2 h-3 w-3" />
                          Replay
                        </Button>
                      )}
                    </div>
                  </div>
                  {e.retry_count > 0 && <p className="mt-1 text-slate-500">Retries: {e.retry_count}</p>}
                  {!!e.payload?.error && (
                    <div className="mt-2 flex items-start gap-2 rounded bg-red-50 p-2 text-red-700">
                      <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
                      <span className="break-all">{String(e.payload.error)}</span>
                    </div>
                  )}
                  <details className="mt-2">
                    <summary className="cursor-pointer text-slate-600">Payload</summary>
                    <pre className="mt-2 max-h-48 overflow-auto rounded bg-slate-50 p-2 text-xs">
                      {JSON.stringify(e.payload, null, 2)}
                    </pre>
                  </details>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
