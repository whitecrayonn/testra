"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { Plus, ChevronRight, Play } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { CardSkeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { LinkButton } from "@/components/ui/link-button";
import { listTestRuns } from "@/features/results/api";
import type { TestRun, PaginationMeta } from "@/types/results";

const statusVariants: Record<string, "neutral" | "info" | "success" | "danger" | "warning"> = {
  pending: "neutral",
  running: "info",
  passed: "success",
  failed: "danger",
  skipped: "warning",
  cancelled: "neutral",
};

const sourceVariants: Record<string, "default" | "info" | "neutral"> = {
  manual: "default",
  ci: "info",
  api: "neutral",
};

export default function TestRunsPage() {
  const [runs, setRuns] = useState<TestRun[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [cursor, setCursor] = useState<string | undefined>(undefined);
  const [error, setError] = useState<string | null>(null);

  const projectId =
    typeof window !== "undefined"
      ? localStorage.getItem("testra_project_id") || ""
      : "";

  const fetchRuns = useCallback(
    async (reset?: boolean) => {
      setLoading(true);
      setError(null);
      try {
        const result = await listTestRuns(projectId, {
          cursor: reset ? undefined : cursor,
        });
        setRuns((prev) => (reset ? result.data : [...prev, ...result.data]));
        setMeta(result.meta);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load runs");
      } finally {
        setLoading(false);
      }
    },
    [projectId, cursor],
  );

  useEffect(() => {
    if (projectId) {
      fetchRuns(true);
    } else {
      setLoading(false);
    }
  }, [projectId, fetchRuns]);

  const loadMore = () => {
    if (meta?.next_cursor) {
      setCursor(meta.next_cursor);
    }
  };

  useEffect(() => {
    if (cursor) {
      fetchRuns(false);
    }
  }, [cursor, fetchRuns]);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Test Runs"
        description="View and manage test execution runs."
        actions={
          <LinkButton href="/dashboard/test-runs/new">
            <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
            New Run
          </LinkButton>
        }
      />

      {error && (
        <div role="alert">
          <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">
            {error}
          </Card>
        </div>
      )}

      {!projectId ? (
        <EmptyState
          icon={Play}
          title="No project selected"
          description="Select a project from the Projects page to view and create test runs."
          action={{ label: "Go to Projects", href: "/dashboard/projects" }}
        />
      ) : loading && runs.length === 0 ? (
        <CardSkeleton count={3} />
      ) : runs.length === 0 ? (
        <EmptyState
          icon={Play}
          title="No test runs yet"
          description="Create your first test run to track execution progress."
          action={{ label: "New Run", href: "/dashboard/test-runs/new" }}
          secondaryAction={{ label: "Go to Projects", href: "/dashboard/projects" }}
        />
      ) : (
        <div className="space-y-3">
          {runs.map((run) => (
            <Link key={run.id} href={`/dashboard/test-runs/${run.id}`} className="group block">
              <Card className="p-4 transition-shadow group-hover:shadow-md">
                <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <h3 className="font-medium text-slate-900">{run.name}</h3>
                    <div className="mt-1 flex flex-wrap items-center gap-2">
                      <Badge variant={statusVariants[run.status] || "neutral"}>{run.status}</Badge>
                      <Badge variant={sourceVariants[run.source] || "neutral"}>{run.source}</Badge>
                      <span className="text-xs text-slate-500">
                        {run.passed} passed · {run.failed} failed · {run.skipped} skipped · {run.total} total
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-xs text-slate-500">
                      {new Date(run.created_at).toLocaleString()}
                    </span>
                    <ChevronRight className="h-4 w-4 text-slate-400" aria-hidden="true" />
                  </div>
                </div>
              </Card>
            </Link>
          ))}

          {meta?.has_more && (
            <div className="flex justify-center pt-4">
              <Button variant="secondary" onClick={loadMore} loading={loading}>
                Load More
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
