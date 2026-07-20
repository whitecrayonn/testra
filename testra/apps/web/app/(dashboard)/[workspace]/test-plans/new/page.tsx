"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { createTestPlan } from "@/features/results/api";

export default function NewTestPlanPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [testCaseIds, setTestCaseIds] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const workspaceId =
    typeof window !== "undefined" ? localStorage.getItem("testra_workspace_id") || "" : "";
  const projectId =
    typeof window !== "undefined" ? localStorage.getItem("testra_project_id") || "" : "";

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      await createTestPlan({
        workspace_id: workspaceId,
        project_id: projectId,
        name,
        description,
        test_case_ids: testCaseIds.split(",").map((s) => s.trim()).filter(Boolean),
      });
      router.push("/dashboard/test-plans");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create test plan");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader title="New Test Plan" description="Create a reusable collection of test cases." />

      {error && (
        <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-1">
          <label htmlFor="name" className="text-sm font-medium text-slate-700">
            Name
          </label>
          <input
            id="name"
            type="text"
            required
            className="w-full rounded-md border border-slate-300 p-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </div>

        <div className="space-y-1">
          <label htmlFor="description" className="text-sm font-medium text-slate-700">
            Description
          </label>
          <textarea
            id="description"
            className="w-full rounded-md border border-slate-300 p-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>

        <div className="space-y-1">
          <label htmlFor="testCaseIds" className="text-sm font-medium text-slate-700">
            Test Case IDs (comma separated)
          </label>
          <input
            id="testCaseIds"
            type="text"
            className="w-full rounded-md border border-slate-300 p-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
            value={testCaseIds}
            onChange={(e) => setTestCaseIds(e.target.value)}
          />
        </div>

        <Button type="submit" loading={submitting}>
          <FileText className="mr-2 h-4 w-4" aria-hidden="true" />
          Create Plan
        </Button>
      </form>
    </div>
  );
}
