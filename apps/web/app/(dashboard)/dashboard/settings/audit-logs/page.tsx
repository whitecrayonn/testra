"use client";

import { useEffect, useState } from "react";
import { ClipboardList, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { listMyAuditEvents } from "@/features/audit/api";
import type { AuditEvent } from "@/types/audit";

function formatTimestamp(iso: string): string {
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

function actionVariant(action: string): "neutral" | "info" | "warning" | "danger" {
  if (action.includes("delete") || action.includes("revoke") || action.includes("disable")) return "danger";
  if (action.includes("update") || action.includes("reset") || action.includes("change")) return "warning";
  if (action.includes("create") || action.includes("login") || action.includes("register")) return "info";
  return "neutral";
}

export default function AuditLogsSettingsPage() {
  const [events, setEvents] = useState<AuditEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [nextCursor, setNextCursor] = useState<string | undefined>(undefined);
  const [hasMore, setHasMore] = useState(false);

  async function loadFirstPage() {
    setLoading(true);
    setError(null);
    try {
      const res = await listMyAuditEvents();
      setEvents(res.data);
      setHasMore(res.meta.has_more);
      setNextCursor(res.meta.next_cursor);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load audit events");
    } finally {
      setLoading(false);
    }
  }

  async function loadMore() {
    if (!nextCursor) return;
    setLoadingMore(true);
    try {
      const res = await listMyAuditEvents(nextCursor);
      setEvents((prev) => [...prev, ...res.data]);
      setHasMore(res.meta.has_more);
      setNextCursor(res.meta.next_cursor);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load more audit events");
    } finally {
      setLoadingMore(false);
    }
  }

  useEffect(() => {
    loadFirstPage();
  }, []);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Audit Logs"
        description="Your own security-relevant activity — logins, password changes, and other account events."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "Audit Logs" },
        ]}
        actions={
          <Button variant="secondary" size="sm" onClick={loadFirstPage} loading={loading}>
            <RefreshCw className="mr-2 h-4 w-4" aria-hidden="true" />
            Refresh
          </Button>
        }
      />

      <Card className="border-blue-100 bg-blue-50 p-4 text-sm text-blue-800">
        This view shows events tied to your own account. An organization-wide audit view for
        owners/admins is planned once audit events carry an organization ID (see the engineering
        backlog, SBL-080).
      </Card>

      {error && (
        <div role="alert">
          <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>
        </div>
      )}

      {!loading && events.length === 0 && !error && (
        <EmptyState
          icon={ClipboardList}
          title="No activity yet"
          description="Security-relevant actions on your account, like logins and password changes, will show up here."
        />
      )}

      {events.length > 0 && (
        <Card className="p-0">
          <ul className="divide-y divide-slate-200">
            {events.map((e) => (
              <li key={e.id} className="flex items-center justify-between gap-3 px-6 py-3">
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <Badge variant={actionVariant(e.action)}>{e.action}</Badge>
                    <span className="truncate text-sm text-slate-700">{e.resource}</span>
                  </div>
                  <p className="mt-1 text-xs text-slate-500">
                    {formatTimestamp(e.created_at)}
                    {e.ip_address ? ` · ${e.ip_address}` : ""}
                  </p>
                </div>
              </li>
            ))}
          </ul>
        </Card>
      )}

      {hasMore && (
        <div className="flex justify-center">
          <Button variant="ghost" size="sm" onClick={loadMore} loading={loadingMore}>
            Load more
          </Button>
        </div>
      )}
    </div>
  );
}
