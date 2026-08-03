"use client";

import { useEffect, useState } from "react";
import { CheckCircle, XCircle, SkipForward, Play } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { getTestCase } from "@/features/testmanagement/api";
import { executeTestRunItem, attachEvidence, getExecutionHistory } from "@/features/results/api";
import type { TestRunItem, StepResult, Evidence, ExecutionHistoryEntry, RunItemStatus } from "@/types/results";
import type { TestCase, TestStep } from "@/types/testmanagement";

const stepStatuses: RunItemStatus[] = ["passed", "failed", "skipped", "blocked", "retest"];

interface StepRunnerProps {
  item: TestRunItem;
  onUpdated?: (item: TestRunItem) => void;
}

export function StepRunner({ item, onUpdated }: StepRunnerProps) {
  const [testCase, setTestCase] = useState<TestCase | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [steps, setSteps] = useState<StepResult[]>([]);
  const [activeStep, setActiveStep] = useState(0);
  const [comment, setComment] = useState(item.comment || "");
  const [startedAt, setStartedAt] = useState<number | null>(null);
  const [history, setHistory] = useState<ExecutionHistoryEntry[]>([]);
  const [evidence, setEvidence] = useState<Evidence[]>([]);

  useEffect(() => {
    if (item.test_case_id) {
      setLoading(true);
      getTestCase(item.test_case_id)
        .then((tc) => setTestCase(tc))
        .finally(() => setLoading(false));
    }
  }, [item.test_case_id]);

  useEffect(() => {
    const initial: StepResult[] = (testCase?.steps || []).map((s, i) => {
      const existing = item.step_results?.find((r) => r.order === i);
      return existing || { order: i, status: "not_executed", comment: "", duration_ms: 0 };
    });
    setSteps(initial);
  }, [testCase, item.step_results]);

  useEffect(() => {
    getExecutionHistory(item.id).then((res) => setHistory(res.data));
  }, [item.id]);

  const start = () => {
    setStartedAt(Date.now());
    setActiveStep(0);
  };

  const setStepStatus = (order: number, status: RunItemStatus) => {
    setSteps((prev) =>
      prev.map((s, i) =>
        i === order
          ? { ...s, status, executed_at: new Date().toISOString() }
          : s,
      ),
    );
    if (order < steps.length - 1) {
      setActiveStep(order + 1);
    }
  };

  const setStepComment = (order: number, c: string) => {
    setSteps((prev) => prev.map((s, i) => (i === order ? { ...s, comment: c } : s)));
  };

  const handleAttachEvidence = async () => {
    const fileName = prompt("Evidence file name");
    if (!fileName) return;
    const res = await attachEvidence(item.id, {
      step_order: activeStep,
      file_name: fileName,
    });
    setEvidence((prev) => [...prev, res.data]);
  };

  const handleSave = async (finalStatus: RunItemStatus) => {
    const durationMs = startedAt ? Date.now() - startedAt : 0;
    setSaving(true);
    try {
      const updated = await executeTestRunItem(item.id, {
        status: finalStatus,
        step_results: steps,
        comment,
        duration_ms: durationMs,
      });
      onUpdated?.(updated as TestRunItem);
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <Card className="p-4 text-sm text-slate-500">Loading test case...</Card>;

  return (
    <div className="space-y-4">
      {testCase && (
        <div className="space-y-1">
          <div className="text-sm text-slate-500">Test case</div>
          <div className="font-medium text-slate-900">{testCase.title}</div>
          {testCase.description && <div className="text-sm text-slate-600">{testCase.description}</div>}
        </div>
      )}

      {startedAt === null ? (
        <Button onClick={start}>
          <Play className="mr-2 h-4 w-4" aria-hidden="true" />
          Start Executing
        </Button>
      ) : (
        <>
          <div className="space-y-2">
            {(testCase?.steps || []).map((step: TestStep, idx: number) => (
              <Card
                key={idx}
                className={`p-3 ${idx === activeStep ? "border-brand-500 ring-1 ring-brand-500" : ""}`}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="flex-1 space-y-1">
                    <div className="text-sm font-medium text-slate-900">
                      Step {step.order + 1}: {step.action}
                    </div>
                    <div className="text-xs text-slate-500">Expected: {step.expected}</div>
                    {step.test_data && <div className="text-xs text-slate-600">Data: {step.test_data}</div>}
                  </div>
                  <div className="flex gap-1">
                    {stepStatuses.map((status) => {
                      const active = steps[idx]?.status === status;
                      return (
                        <Button
                          key={status}
                          variant={active ? "primary" : "secondary"}
                          size="sm"
                          onClick={() => setStepStatus(idx, status)}
                        >
                          {status}
                        </Button>
                      );
                    })}
                  </div>
                </div>
                <textarea
                  className="mt-2 min-h-[60px] w-full rounded-md border border-slate-300 p-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
                  placeholder="Step comment"
                  value={steps[idx]?.comment || ""}
                  onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setStepComment(idx, e.target.value)}
                />
              </Card>
            ))}
          </div>

          <textarea
            className="min-h-[80px] w-full rounded-md border border-slate-300 p-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
            placeholder="Overall comment for this item"
            value={comment}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setComment(e.target.value)}
          />

          <div className="flex flex-wrap gap-2">
            <Button onClick={() => handleSave("passed")} disabled={saving}>
              <CheckCircle className="mr-2 h-4 w-4" aria-hidden="true" />
              Save as Passed
            </Button>
            <Button variant="secondary" onClick={() => handleSave("failed")} disabled={saving}>
              <XCircle className="mr-2 h-4 w-4" aria-hidden="true" />
              Fail
            </Button>
            <Button variant="secondary" onClick={() => handleSave("blocked")} disabled={saving}>
              <SkipForward className="mr-2 h-4 w-4" aria-hidden="true" />
              Block
            </Button>
            <Button variant="secondary" onClick={handleAttachEvidence}>
              Attach Evidence
            </Button>
          </div>

          {evidence.length > 0 && (
            <div className="space-y-1">
              <div className="text-sm font-medium text-slate-900">Evidence</div>
              {evidence.map((e) => (
                <div key={e.id} className="text-sm text-slate-600">
                  {e.file_name}
                </div>
              ))}
            </div>
          )}
        </>
      )}

      {history.length > 0 && (
        <div className="space-y-2">
          <div className="text-sm font-medium text-slate-900">Execution History</div>
          {history.map((h) => (
            <Card key={h.id} className="p-3 text-sm">
              <div className="flex items-center gap-2">
                <Badge variant="neutral">{h.status}</Badge>
                <span className="text-xs text-slate-500">{new Date(h.created_at).toLocaleString()}</span>
              </div>
              {h.comment && <div className="mt-1 text-slate-700">{h.comment}</div>}
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
