"use client";

import { useState, useEffect, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Play, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { LinkButton } from "@/components/ui/link-button";
import { PageHeader } from "@/components/ui/page-header";
import { getTestPlan, getTestPlanItems, deleteTestPlan, createRunFromPlan } from "@/features/results/api";
import type { TestPlan, TestPlanItem } from "@/types/results";

export default function TestPlanDetailPage() {
  const params = useParams();
  const router = useRouter();
  const planId = params.id as string;
  const [plan, setPlan] = useState<TestPlan | null>(null);
  const [items, setItems] = useState<TestPlanItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [starting, setStarting] = useState(false);

  const fetchPlan = useCallback(async () => {
    try {
      const p = await getTestPlan(planId);
      setPlan(p.data);
      const i = await getTestPlanItems(planId);
      setItems(i.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load test plan");
    } finally {
      setLoading(false);
    }
  }, [planId]);

  useEffect(() => {
    fetchPlan();
  }, [fetchPlan]);

  const handleStartRun = async () => {
    setStarting(true);
    try {
      await createRunFromPlan(planId);
      router.push("/dashboard/test-runs");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to start run");
    } finally {
      setStarting(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Delete this test plan?")) return;
    try {
      await deleteTestPlan(planId);
      router.push("/dashboard/test-plans");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete test plan");
    }
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <PageHeader title="Test Plan" description="Loading..." />
        <Card className="p-8 text-center text-slate-500">Loading...</Card>
      </div>
    );
  }

  if (error || !plan) {
    return (
      <div className="space-y-4">
        <LinkButton href="/dashboard/test-plans" variant="ghost" size="sm">
          <ArrowLeft className="mr-2 h-4 w-4" aria-hidden="true" />
          Back to Plans
        </LinkButton>
        <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">
          {error || "Plan not found"}
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={plan.name}
        description={plan.description || "Test plan details"}
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Test Plans", href: "/dashboard/test-plans" },
          { label: plan.name },
        ]}
        actions={
          <div className="flex gap-2">
            <Button onClick={handleStartRun} loading={starting}>
              <Play className="mr-2 h-4 w-4" aria-hidden="true" />
              Create Run
            </Button>
            <Button variant="secondary" onClick={handleDelete}>
              <Trash2 className="mr-2 h-4 w-4" aria-hidden="true" />
              Delete
            </Button>
          </div>
        }
      />

      <div className="flex items-center gap-2">
        <Badge variant={plan.status === "active" ? "success" : "neutral"}>{plan.status}</Badge>
      </div>

      <div className="space-y-3">
        <h2 className="text-lg font-semibold text-slate-900">Test Cases ({items.length})</h2>
        {items.length === 0 ? (
          <Card className="p-6 text-center text-slate-500">No test cases in this plan.</Card>
        ) : (
          <div className="space-y-2">
            {items.map((item) => (
              <Card key={item.id} className="p-3 text-sm text-slate-700">
                <div className="flex items-center justify-between">
                  <span>Test case ID: {item.test_case_id}</span>
                  <span className="text-xs text-slate-500">Order: {item.sort_order}</span>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
