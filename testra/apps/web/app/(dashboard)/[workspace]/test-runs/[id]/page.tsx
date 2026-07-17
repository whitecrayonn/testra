"use client";

import { useState, useEffect, useCallback, useMemo } from "react";
import { useParams } from "next/navigation";
import { ArrowLeft, Play, CheckCircle, XCircle, Clock, SkipForward } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { LinkButton } from "@/components/ui/link-button";
import { PageHeader } from "@/components/ui/page-header";
import { getToken } from "@/lib/api";
import { getTestRun, listTestRunItems, updateTestRunStatus } from "@/features/results/api";
import type { TestRun, TestRunItem, RunProgressEvent } from "@/types/results";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

function isNilUUID(id: string | null | undefined): boolean {
  return !id || id === "00000000-0000-0000-0000-000000000000";
}

function isTerminalRunStatus(status: string): boolean {
  return ["passed", "failed", "skipped", "cancelled"].includes(status);
}

const statusVariants: Record<string, "neutral" | "info" | "success" | "danger" | "warning"> = {
  pending: "neutral",
  running: "info",
  passed: "success",
  failed: "danger",
  skipped: "warning",
  blocked: "warning",
  cancelled: "neutral",
};

const itemStatusIcons: Record<string, React.ReactNode> = {
  pending: <Clock className="h-4 w-4 text-slate-400" aria-hidden="true" />,
  running: <Play className="h-4 w-4 text-brand-600" aria-hidden="true" />,
  passed: <CheckCircle className="h-4 w-4 text-green-600" aria-hidden="true" />,
  failed: <XCircle className="h-4 w-4 text-red-600" aria-hidden="true" />,
  skipped: <SkipForward className="h-4 w-4 text-yellow-600" aria-hidden="true" />,
  blocked: <Clock className="h-4 w-4 text-orange-600" aria-hidden="true" />,
};

export default function TestRunDetailPage() {
  const params = useParams();
  const runId = params.id as string;
  const [run, setRun] = useState<TestRun | null>(null);
  const [items, setItems] = useState<TestRunItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [progress, setProgress] = useState<RunProgressEvent | null>(null);

  const isActive = useMemo(
    () => run?.status === "pending" || run?.status === "running",
    [run?.status],
  );

  const fetchRun = useCallback(async () => {
    try {
      const r = await getTestRun(runId);
      setRun(r);
      const itemResults = await listTestRunItems(runId);
      setItems(itemResults);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load run");
    } finally {
      setLoading(false);
    }
  }, [runId]);

  useEffect(() => {
    fetchRun();
  }, [fetchRun]);

  useEffect(() => {
    if (!isActive) return;

    const token = getToken();
    if (!token) return;

    const eventSource = new EventSource(
      `${API_URL}/api/v1/test-runs/${runId}/stream?access_token=${encodeURIComponent(token)}`,
    );

    eventSource.onmessage = (e) => {
      const event: RunProgressEvent = JSON.parse(e.data);
      setProgress(event);

      setRun((prev) => {
        if (!prev) return null;
        const next = {
          ...prev,
          passed: event.passed,
          failed: event.failed,
          skipped: event.skipped,
          blocked: event.blocked,
          total: event.total,
        };
        // When the server sends a run-level update it omits an item_id.
        if (isNilUUID(event.item_id)) {
          return { ...next, status: event.status as TestRun["status"] };
        }
        return next;
      });

      if (!isNilUUID(event.item_id)) {
        setItems((prev) =>
          prev.map((item) =>
            item.id === event.item_id
              ? { ...item, status: event.status as TestRunItem["status"] }
              : item,
          ),
        );
      }

      if (isNilUUID(event.item_id) && isTerminalRunStatus(event.status)) {
        eventSource.close();
        fetchRun();
      }
    };

    eventSource.onerror = () => {
      eventSource.close();
    };

    return () => {
      eventSource.close();
    };
  }, [runId, isActive, fetchRun]);

  const handleStartRun = async () => {
    try {
      const updated = await updateTestRunStatus(runId, "running");
      setRun(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to start run");
    }
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <PageHeader title="Test Run" description="Loading run details..." />
        <Card className="p-8 text-center text-slate-500">Loading...</Card>
      </div>
    );
  }

  if (error || !run) {
    return (
      <div className="space-y-4">
        <LinkButton href="/dashboard/test-runs" variant="ghost" size="sm">
          <ArrowLeft className="mr-2 h-4 w-4" aria-hidden="true" />
          Back to Runs
        </LinkButton>
        <div role="alert">
          <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">
            {error || "Run not found"}
          </Card>
        </div>
      </div>
    );
  }

  const displayProgress = progress?.progress ?? (run.total > 0 ? (run.passed + run.failed + run.skipped + run.blocked) / run.total : 0);

  return (
    <div className="space-y-6">
      <PageHeader
        title={run.name}
        description={`${run.source} · ${new Date(run.created_at).toLocaleString()}`}
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Runs", href: "/dashboard/test-runs" },
          { label: run.name },
        ]}
        actions={
          run.status === "pending" && (
            <Button onClick={handleStartRun}>
              <Play className="mr-2 h-4 w-4" aria-hidden="true" />
              Start Run
            </Button>
          )
        }
      />

      <div className="flex items-center gap-2">
        <Badge variant={statusVariants[run.status] || "neutral"}>{run.status}</Badge>
        <Badge variant={run.source === "manual" ? "default" : run.source === "ci" ? "info" : "neutral"}>
          {run.source}
        </Badge>
      </div>

      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <Card className="p-4">
          <div className="text-2xl font-bold text-green-600">{progress?.passed ?? run.passed}</div>
          <div className="text-xs text-slate-500">Passed</div>
        </Card>
        <Card className="p-4">
          <div className="text-2xl font-bold text-red-600">{progress?.failed ?? run.failed}</div>
          <div className="text-xs text-slate-500">Failed</div>
        </Card>
        <Card className="p-4">
          <div className="text-2xl font-bold text-yellow-600">{progress?.skipped ?? run.skipped}</div>
          <div className="text-xs text-slate-500">Skipped</div>
        </Card>
        <Card className="p-4">
          <div className="text-2xl font-bold text-slate-900">{progress?.total ?? run.total}</div>
          <div className="text-xs text-slate-500">Total</div>
        </Card>
      </div>

      {run.status === "running" && (
        <div className="space-y-2">
          <div className="flex justify-between text-xs text-slate-600">
            <span>Progress</span>
            <span>{Math.round(displayProgress * 100)}%</span>
          </div>
          <div
            className="h-2 overflow-hidden rounded-full bg-slate-200"
            role="progressbar"
            aria-valuenow={Math.round(displayProgress * 100)}
            aria-valuemin={0}
            aria-valuemax={100}
          >
            <div
              className="h-full bg-brand-600 transition-all duration-300"
              style={{ width: `${displayProgress * 100}%` }}
            />
          </div>
        </div>
      )}

      <div className="space-y-3">
        <h2 className="text-lg font-semibold text-slate-900">Test Items</h2>
        {items.length === 0 ? (
          <Card className="p-6 text-center text-slate-500">No items in this run.</Card>
        ) : (
          <div className="space-y-2">
            {items.map((item) => (
              <Card key={item.id} className="p-3">
                <div className="flex items-center gap-3">
                  {itemStatusIcons[item.status] || <Clock className="h-4 w-4 text-slate-400" aria-hidden="true" />}
                  <div className="flex-1">
                    <div className="text-sm font-medium text-slate-900">{item.title}</div>
                    {item.error_message && (
                      <div className="mt-1 font-mono text-xs text-red-600">{item.error_message}</div>
                    )}
                  </div>
                  <Badge variant={statusVariants[item.status] || "neutral"}>{item.status}</Badge>
                  {item.duration_ms > 0 && (
                    <span className="text-xs text-slate-500">{(item.duration_ms / 1000).toFixed(2)}s</span>
                  )}
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
