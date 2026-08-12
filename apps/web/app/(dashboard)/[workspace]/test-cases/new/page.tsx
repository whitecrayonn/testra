"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Plus, Trash2, TestTube } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { LinkButton } from "@/components/ui/link-button";
import { createTestCase } from "@/features/testmanagement/api";

interface StepForm {
  action: string;
  expected: string;
  test_data: string;
}

export default function NewTestCasePage() {
  const router = useRouter();
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [projectId, setProjectId] = useState("");
  const [workspaceId, setWorkspaceId] = useState("");

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [preconditions, setPreconditions] = useState("");
  const [status, setStatus] = useState("draft");
  const [priority, setPriority] = useState("medium");
  const [tagsInput, setTagsInput] = useState("");
  const [steps, setSteps] = useState<StepForm[]>([]);

  useEffect(() => {
    if (typeof window !== "undefined") {
      setProjectId(localStorage.getItem("testra_project_id") || "");
      setWorkspaceId(localStorage.getItem("testra_workspace_id") || "");
    }
  }, []);

  function addStep() {
    setSteps([...steps, { action: "", expected: "", test_data: "" }]);
  }

  function updateStep(index: number, field: keyof StepForm, value: string) {
    const updated = [...steps];
    updated[index] = { ...updated[index], [field]: value };
    setSteps(updated);
  }

  function removeStep(index: number) {
    setSteps(steps.filter((_, i) => i !== index));
  }

  async function handleCreate() {
    if (!title.trim()) {
      setError("Title is required");
      return;
    }
    if (!projectId || !workspaceId) {
      setError("No project or workspace selected. Select a project first.");
      return;
    }

    setSaving(true);
    setError(null);
    try {
      const tags = tagsInput
        .split(",")
        .map((t) => t.trim())
        .filter(Boolean);
      const tc = await createTestCase({
        workspace_id: workspaceId,
        project_id: projectId,
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
      router.push(`/dashboard/test-cases/${tc.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create test case");
    } finally {
      setSaving(false);
    }
  }

  if (!projectId) {
    return (
      <div className="space-y-6">
        <LinkButton href="/dashboard/test-cases" variant="ghost" size="sm">
          <ArrowLeft className="mr-2 h-4 w-4" aria-hidden="true" />
          Back
        </LinkButton>
        <EmptyState
          icon={TestTube}
          title="No project selected"
          description="Select a project in Projects before creating a test case."
          action={{ label: "Go to Projects", href: "/dashboard/projects" }}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="New Test Case"
        description="Add a new test case to your repository."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Test Cases", href: "/dashboard/test-cases" },
          { label: "New" },
        ]}
        actions={
          <Button onClick={handleCreate} loading={saving}>
            Create Test Case
          </Button>
        }
      />

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
          <Input
            id="tc-title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g. Login with valid credentials"
          />
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
    </div>
  );
}
