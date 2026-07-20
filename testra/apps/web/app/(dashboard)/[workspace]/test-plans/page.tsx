"use client";

import { useState, useEffect, useCallback } from "react";
import { Plus, ChevronRight, FileText } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { CardSkeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { LinkButton } from "@/components/ui/link-button";
import { listTestPlans, createRunFromPlan } from "@/features/results/api";
import type { TestPlan, PaginationMeta } from "@/types/results";

const statusVariants: Record<string, "neutral" | "info" | "success" | "danger" | "warning"> = {
  active: "success",
  archived: "neutral",
};

export default function TestPlansPage() {
  const [plans, setPlans] = useState<TestPlan[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [cursor, setCursor] = useState<string | undefined>(undefined);
  const [error, setError] = useState<string | null>(null);
  const [startingId, setStartingId] = useState<string | null>(null);

  const projectId =
    typeof window !== "undefined" ? localStorage.getItem("testra_project_id") || "" : "";

  const fetchPlans = useCallback(
    async (reset?: boolean) => {
      setLoading(true);
      setError(null);
      try {
        const result = await listTestPlans(projectId, {
          cursor: reset ? undefined : cursor,
        });
        setPlans((prev) => (reset ? result.data : [...prev, ...result.data]));
        setMeta(result.meta);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load test plans");
      } finally {
        setLoading(false);
      }
    },
    [projectId, cursor],
  );

  useEffect(() => {
    if (projectId) {
      fetchPlans(true);
    } else {
      setLoading(false);
    }
  }, [projectId, fetchPlans]);

  useEffect(() => {
    if (cursor) {
      fetchPlans(false);
    }
  }, [cursor, fetchPlans]);

  const loadMore = () => {
    if (meta?.next_cursor) {
      setCursor(meta.next_cursor);
    }
  };

  const handleCreateRun = async (e: React.MouseEvent, planId: string) => {
    e.preventDefault();
    e.stopPropagation();
    setStartingId(planId);
    try {
      await createRunFromPlan(planId);
      window.location.href = "/dashboard/test-runs";
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create run");
    } finally {
      setStartingId(null);
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Test Plans"
        description="Reusable collections of test cases that can be turned into runs."
        actions={
          <LinkButton href="/dashboard/test-plans/new">
            <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
            New Plan
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
          icon={FileText}
          title="No project selected"
          description="Select a project from the Projects page to view and create test plans."
          action={{ label: "Go to Projects", href: "/dashboard/projects" }}
        />
      ) : loading && plans.length === 0 ? (
        <CardSkeleton count={3} />
      ) : plans.length === 0 ? (
        <EmptyState
          icon={FileText}
          title="No test plans yet"
          description="Create your first test plan to organize reusable test cases."
          action={{ label: "New Plan", href: "/dashboard/test-plans/new" }}
        />
      ) : (
        <div className="space-y-3">
          {plans.map((plan) => (
            <Link key={plan.id} href={`/dashboard/test-plans/${plan.id}`} className="group block">
              <Card className="p-4 transition-shadow group-hover:shadow-md">
                <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <h3 className="font-medium text-slate-900">{plan.name}</h3>
                    <div className="mt-1 flex flex-wrap items-center gap-2">
                      <Badge variant={statusVariants[plan.status] || "neutral"}>{plan.status}</Badge>
                      <span className="text-xs text-slate-500">{plan.description || "No description"}</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Button
                      variant="secondary"
                      size="sm"
                      loading={startingId === plan.id}
                      onClick={(e) => handleCreateRun(e, plan.id)}
                    >
                      Create Run
                    </Button>
                    <span className="text-xs text-slate-500">
                      {new Date(plan.created_at).toLocaleString()}
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
