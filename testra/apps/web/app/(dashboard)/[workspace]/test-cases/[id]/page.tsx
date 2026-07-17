"use client";

import { useState, useEffect } from "react";
import { useParams } from "next/navigation";
import { ArrowLeft, Plus, Trash2, Save } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { LinkButton } from "@/components/ui/link-button";
import { PageHeader } from "@/components/ui/page-header";
import { getTestCase, updateTestCase, deleteTestCase, listTestCaseVersions } from "@/features/testmanagement/api";
import type { TestCase, TestCaseVersion, TestStep } from "@/types/testmanagement";

const statusVariants: Record<string, "neutral" | "success" | "danger"> = {
  draft: "neutral",
  active: "success",
  deprecated: "danger",
};

const priorityVariants: Record<string, "neutral" | "info" | "warning" | "danger"> = {
  low: "neutral",
  medium: "info",
  high: "warning",
  critical: "danger",
};

export default function TestCaseDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const [tc, setTc] = useState<TestCase | null>(null);
  const [versions, setVersions] = useState<TestCaseVersion[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showVersions, setShowVersions] = useState(false);

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [preconditions, setPreconditions] = useState("");
  const [status, setStatus] = useState("draft");
  const [priority, setPriority] = useState("medium");
  const [tagsInput, setTagsInput] = useState("");
  const [steps, setSteps] = useState<TestStep[]>([]);

  useEffect(() => {
    async function load() {
      try {
        const testCase = await getTestCase(id);
        setTc(testCase);
        setTitle(testCase.title);
        setDescription(testCase.description);
        setPreconditions(testCase.preconditions);
        setStatus(testCase.status);
        setPriority(testCase.priority);
        setTagsInput(testCase.tags.join(", "));
        setSteps(testCase.steps);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load test case");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [id]);

  async function loadVersions() {
    try {
      const v = await listTestCaseVersions(id);
      setVersions(v);
      setShowVersions(true);
    } catch {
      setError("Failed to load version history");
    }
  }

  async function handleSave() {
    setSaving(true);
    setError(null);
    try {
      const tags = tagsInput
        .split(",")
        .map((t) => t.trim())
        .filter(Boolean);
      const updated = await updateTestCase(id, {
        title,
        description,
        preconditions,
        status,
        priority,
        tags,
        steps: steps.map((s) => ({
          action: s.action,
          expected: s.expected,
          test_data: s.test_data,
        })),
      });
      setTc(updated);
      setTitle(updated.title);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    if (!confirm("Are you sure you want to delete this test case?")) return;
    try {
      await deleteTestCase(id);
      window.location.href = "/dashboard/test-cases";
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete");
    }
  }

  function addStep() {
    setSteps([...steps, { order: steps.length + 1, action: "", expected: "", test_data: "" }]);
  }

  function updateStep(index: number, field: keyof TestStep, value: string) {
    const updated = [...steps];
    updated[index] = { ...updated[index], [field]: value };
    setSteps(updated);
  }

  function removeStep(index: number) {
    setSteps(steps.filter((_, i) => i !== index).map((s, i) => ({ ...s, order: i + 1 })));
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <PageHeader title="Test Case" description="Loading test case details..." />
        <Card className="p-8 text-center text-slate-500">Loading...</Card>
      </div>
    );
  }

  if (!tc) {
    return (
      <div className="space-y-4">
        <LinkButton href="/dashboard/test-cases" variant="ghost" size="sm">
          <ArrowLeft className="mr-2 h-4 w-4" aria-hidden="true" />
          Back to Test Cases
        </LinkButton>
        <div role="alert">
          <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">
            {error || "Test case not found."}
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={tc.title}
        description={`Version ${tc.version}`}
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Test Cases", href: "/dashboard/test-cases" },
          { label: tc.title },
        ]}
        actions={
          <div className="flex flex-wrap gap-2">
            <Button variant="ghost" size="sm" onClick={loadVersions}>
              Version History
            </Button>
            <Button variant="danger" size="sm" onClick={handleDelete}>
              <Trash2 className="mr-2 h-4 w-4" aria-hidden="true" />
              Delete
            </Button>
            <Button onClick={handleSave} loading={saving} size="sm">
              <Save className="mr-2 h-4 w-4" aria-hidden="true" />
              Save
            </Button>
          </div>
        }
      />

      <div className="flex flex-wrap gap-2">
        <Badge variant={statusVariants[tc.status] || "neutral"}>{tc.status}</Badge>
        <Badge variant={priorityVariants[tc.priority] || "neutral"}>{tc.priority}</Badge>
      </div>

      {error && (
        <div role="alert">
          <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">
            {error}
          </Card>
        </div>
      )}

      <Card className="space-y-4 p-6">
        <div>
          <label htmlFor="tc-title" className="mb-1 block text-sm font-medium text-slate-700">
            Title
          </label>
          <Input id="tc-title" value={title} onChange={(e) => setTitle(e.target.value)} />
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label htmlFor="tc-status" className="mb-1 block text-sm font-medium text-slate-700">
              Status
            </label>
            <select
              id="tc-status"
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
            >
              <option value="draft">Draft</option>
              <option value="active">Active</option>
              <option value="deprecated">Deprecated</option>
            </select>
          </div>
          <div>
            <label htmlFor="tc-priority" className="mb-1 block text-sm font-medium text-slate-700">
              Priority
            </label>
            <select
              id="tc-priority"
              value={priority}
              onChange={(e) => setPriority(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
            >
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </select>
          </div>
        </div>

        <div>
          <label htmlFor="tc-description" className="mb-1 block text-sm font-medium text-slate-700">
            Description
          </label>
          <textarea
            id="tc-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
          />
        </div>

        <div>
          <label htmlFor="tc-preconditions" className="mb-1 block text-sm font-medium text-slate-700">
            Preconditions
          </label>
          <textarea
            id="tc-preconditions"
            value={preconditions}
            onChange={(e) => setPreconditions(e.target.value)}
            rows={2}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
          />
        </div>

        <div>
          <label htmlFor="tc-tags" className="mb-1 block text-sm font-medium text-slate-700">
            Tags (comma-separated)
          </label>
          <Input
            id="tc-tags"
            value={tagsInput}
            onChange={(e) => setTagsInput(e.target.value)}
            placeholder="smoke, regression, api"
          />
        </div>
      </Card>

      <Card className="space-y-4 p-6">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-900">Test Steps</h2>
          <Button variant="secondary" size="sm" onClick={addStep}>
            <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
            Add Step
          </Button>
        </div>

        {steps.length === 0 ? (
          <p className="py-4 text-center text-sm text-slate-500">
            No steps yet. Add your first test step.
          </p>
        ) : (
          <div className="space-y-3">
            {steps.map((step, i) => (
              <div
                key={i}
                className="flex gap-3 rounded-lg border border-slate-200 p-3"
              >
                <span className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-brand-100 text-sm font-medium text-brand-700">
                  {i + 1}
                </span>
                <div className="flex-1 space-y-2">
                  <Input
                    placeholder="Action"
                    value={step.action}
                    onChange={(e) => updateStep(i, "action", e.target.value)}
                    aria-label={`Step ${i + 1} action`}
                  />
                  <Input
                    placeholder="Expected result"
                    value={step.expected}
                    onChange={(e) => updateStep(i, "expected", e.target.value)}
                    aria-label={`Step ${i + 1} expected result`}
                  />
                  <Input
                    placeholder="Test data (optional)"
                    value={step.test_data}
                    onChange={(e) => updateStep(i, "test_data", e.target.value)}
                    aria-label={`Step ${i + 1} test data`}
                  />
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => removeStep(i)}
                  className="text-red-600 hover:bg-red-50"
                  aria-label={`Remove step ${i + 1}`}
                >
                  <Trash2 className="h-4 w-4" aria-hidden="true" />
                </Button>
              </div>
            ))}
          </div>
        )}
      </Card>

      {showVersions && (
        <Card className="space-y-3 p-6">
          <h2 className="text-lg font-semibold text-slate-900">Version History</h2>
          {versions.length === 0 ? (
            <p className="text-sm text-slate-500">No previous versions.</p>
          ) : (
            <div className="space-y-2">
              {versions.map((v) => (
                <div
                  key={v.id}
                  className="flex items-center justify-between rounded-lg border border-slate-200 p-3"
                >
                  <div>
                    <span className="font-medium text-slate-900">v{v.version}</span>
                    <span className="ml-2 text-sm text-slate-500">{v.title}</span>
                  </div>
                  <span className="text-xs text-slate-400">{new Date(v.created_at).toLocaleString()}</span>
                </div>
              ))}
            </div>
          )}
        </Card>
      )}
    </div>
  );
}
