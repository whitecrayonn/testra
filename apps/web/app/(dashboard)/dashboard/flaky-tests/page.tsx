"use client";

import { useEffect, useState } from "react";
import { Brain, Loader2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { listPredictions, listClusters } from "@/features/intelligence/api";
import type { FlakyPrediction, FailureCluster } from "@/types/intelligence";

export default function FlakyTestsPage() {
  const [predictions, setPredictions] = useState<FlakyPrediction[]>([]);
  const [clusters, setClusters] = useState<FailureCluster[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const workspaceId = localStorage.getItem("testra_workspace_id") || "";
    if (!workspaceId) {
      setLoading(false);
      setError("No workspace selected");
      return;
    }
    Promise.all([listPredictions(workspaceId), listClusters(workspaceId)])
      .then(([p, c]) => {
        setPredictions(p);
        setClusters(c);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load intelligence"))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Flaky tests"
        description="Predictions and failure clusters generated from your test history."
      />

      {error && <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>}

      {loading ? (
        <Card className="flex h-40 items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-slate-400" />
        </Card>
      ) : (
        <>
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Brain className="h-5 w-5 text-brand-600" />
                Flaky predictions
              </CardTitle>
            </CardHeader>
            <CardContent>
              {predictions.length === 0 ? (
                <EmptyState
                  icon={Brain}
                  title="No predictions yet"
                  description="Run more tests or trigger a prediction from the test case page."
                />
              ) : (
                <div className="space-y-3">
                  {predictions.map((p) => (
                    <div key={p.id} className="rounded-lg border border-slate-200 p-4">
                      <div className="flex items-center justify-between">
                        <h3 className="font-semibold text-slate-900">{p.test_case_title}</h3>
                        <span className="text-sm font-medium text-slate-600">{Math.round(p.flakiness_score * 100)}% flaky</span>
                      </div>
                      <p className="mt-1 text-sm text-slate-500">{p.explanation}</p>
                      <p className="mt-1 text-xs text-slate-400">Recommended: {p.recommended_action}</p>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Failure clusters</CardTitle>
            </CardHeader>
            <CardContent>
              {clusters.length === 0 ? (
                <EmptyState icon={Brain} title="No clusters yet" description="Clusters appear as failures are classified." />
              ) : (
                <div className="space-y-3">
                  {clusters.map((c) => (
                    <div key={c.id} className="flex items-center justify-between rounded-lg border border-slate-200 p-4">
                      <div>
                        <h3 className="font-semibold text-slate-900">{c.label}</h3>
                        <p className="text-xs text-slate-500">{c.signature}</p>
                      </div>
                      <span className="rounded-full bg-slate-100 px-2 py-1 text-xs font-medium text-slate-700">{c.count}</span>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
