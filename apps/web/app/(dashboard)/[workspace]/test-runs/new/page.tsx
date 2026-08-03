"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Play } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { LinkButton } from "@/components/ui/link-button";
import { createTestRun } from "@/features/results/api";

export default function NewTestRunPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [testCaseIds, setTestCaseIds] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [workspaceId, setWorkspaceId] = useState("");
  const [projectId, setProjectId] = useState("");

  useEffect(() => {
    if (typeof window !== "undefined") {
      setWorkspaceId(localStorage.getItem("testra_workspace_id") || "");
      setProjectId(localStorage.getItem("testra_project_id") || "");
    }
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const ids = testCaseIds
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean);
      const run = await createTestRun({
        workspace_id: workspaceId,
        project_id: projectId,
        name,
        test_case_ids: ids,
        source: "manual",
      });
      router.push(`/dashboard/test-runs/${run.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create run");
    } finally {
      setLoading(false);
    }
  };

  if (!projectId) {
    return (
      <div className="space-y-6">
        <LinkButton href="/dashboard/test-runs" variant="ghost" size="sm">
          <ArrowLeft className="mr-2 h-4 w-4" aria-hidden="true" />
          Back to Runs
        </LinkButton>
        <EmptyState
          icon={Play}
          title="No project selected"
          description="Select a project from the Projects page before creating a test run."
          action={{ label: "Go to Projects", href: "/dashboard/projects" }}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="New Test Run"
        description="Create a manual test run from selected test cases."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Runs", href: "/dashboard/test-runs" },
          { label: "New" },
        ]}
      />

      <Card className="p-6">
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="run-name" className="block text-sm font-medium text-slate-700 mb-1">
              Name
            </label>
            <Input
              id="run-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Nightly Regression"
              required
            />
          </div>
          <div>
            <label htmlFor="run-case-ids" className="block text-sm font-medium text-slate-700 mb-1">
              Test Case IDs (comma-separated)
            </label>
            <Input
              id="run-case-ids"
              value={testCaseIds}
              onChange={(e) => setTestCaseIds(e.target.value)}
              placeholder="uuid1, uuid2, uuid3"
            />
          </div>
          {error && (
            <div role="alert" className="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
              {error}
            </div>
          )}
          <div className="flex gap-2">
            <Button type="submit" loading={loading}>
              Create Run
            </Button>
            <LinkButton href="/dashboard/test-runs" variant="secondary">Cancel</LinkButton>
          </div>
        </form>
      </Card>
    </div>
  );
}
