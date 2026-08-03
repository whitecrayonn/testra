"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { ArrowLeft, Upload, FileJson, CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { LinkButton } from "@/components/ui/link-button";
import { generateFromSpec } from "@/features/testgen/api";
import type { GenerateFromSpecResponse } from "@/types/testgen";

const priorityVariants: Record<string, "neutral" | "info" | "warning" | "danger"> = {
  low: "neutral",
  medium: "info",
  high: "warning",
  critical: "danger",
};

export default function GenerateTestCasesPage() {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [projectId, setProjectId] = useState("");
  const [workspaceId, setWorkspaceId] = useState("");

  const [specText, setSpecText] = useState("");
  const [specFilename, setSpecFilename] = useState("");
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<GenerateFromSpecResponse | null>(null);

  useEffect(() => {
    if (typeof window !== "undefined") {
      setProjectId(localStorage.getItem("testra_project_id") || "");
      setWorkspaceId(localStorage.getItem("testra_workspace_id") || "");
    }
  }, []);

  function handleFilePicked(file: File) {
    setSpecFilename(file.name);
    setError(null);
    const reader = new FileReader();
    reader.onload = () => {
      setSpecText(typeof reader.result === "string" ? reader.result : "");
    };
    reader.onerror = () => setError("Failed to read the selected file.");
    reader.readAsText(file);
  }

  async function handleGenerate() {
    if (!projectId || !workspaceId) {
      setError("No project or workspace selected. Select a project first.");
      return;
    }
    if (!specText.trim()) {
      setError("Paste or upload an OpenAPI spec first.");
      return;
    }

    let spec: Record<string, unknown>;
    try {
      spec = JSON.parse(specText);
    } catch {
      setError(
        "Couldn't parse that as JSON. Only JSON OpenAPI specs are supported (convert YAML to JSON first, e.g. with yq).",
      );
      return;
    }

    setGenerating(true);
    setError(null);
    setResult(null);
    try {
      const response = await generateFromSpec({
        workspace_id: workspaceId,
        project_id: projectId,
        spec_filename: specFilename || undefined,
        spec,
      });
      setResult(response);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to generate test cases");
    } finally {
      setGenerating(false);
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
          icon={FileJson}
          title="No project selected"
          description="Select a project in Projects before generating test cases."
          action={{ label: "Go to Projects", href: "/dashboard/projects" }}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Generate Test Cases"
        description="Create draft test cases from an OpenAPI spec — deterministic and rule-based, no AI/ML involved."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Test Cases", href: "/dashboard/test-cases" },
          { label: "Generate" },
        ]}
        actions={
          <Button onClick={handleGenerate} loading={generating}>
            <FileJson className="mr-2 h-4 w-4" aria-hidden="true" />
            Generate
          </Button>
        }
      />

      {error && (
        <div role="alert">
          <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>
        </div>
      )}

      {!result && (
        <Card className="space-y-4 p-6">
          <div>
            <p className="text-sm text-slate-600">
              Every draft case this creates has status <strong>pending_review</strong> and is
              tagged as generated. Review and approve each one — via this page&apos;s results below,
              or from the test case list — before it counts toward coverage.
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <Button
              type="button"
              variant="secondary"
              onClick={() => fileInputRef.current?.click()}
            >
              <Upload className="mr-2 h-4 w-4" aria-hidden="true" />
              Upload spec (.json)
            </Button>
            <input
              ref={fileInputRef}
              type="file"
              accept=".json,application/json"
              className="hidden"
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (file) handleFilePicked(file);
                e.target.value = "";
              }}
            />
            {specFilename && <span className="text-sm text-slate-500">{specFilename}</span>}
          </div>

          <div>
            <label htmlFor="spec-text" className="mb-1 block text-sm font-medium text-slate-700">
              Or paste the OpenAPI spec JSON
            </label>
            <textarea
              id="spec-text"
              value={specText}
              onChange={(e) => {
                setSpecText(e.target.value);
                if (specFilename) setSpecFilename("");
              }}
              rows={14}
              placeholder='{"openapi": "3.0.3", "paths": { ... } }'
              spellCheck={false}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 font-mono text-xs text-slate-900 focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
            />
            <p className="mt-1 text-xs text-slate-500">
              JSON only (not YAML) — this keeps the generator dependency-free. Convert YAML to JSON
              first if needed.
            </p>
          </div>
        </Card>
      )}

      {result && (
        <>
          <Card className="space-y-3 p-6">
            <div className="flex items-center gap-2">
              <CheckCircle2 className="h-5 w-5 text-green-600" aria-hidden="true" />
              <h2 className="text-lg font-semibold text-slate-900">Generation complete</h2>
            </div>
            <p className="text-sm text-slate-600">
              Created <strong>{result.run.case_count}</strong> draft test case
              {result.run.case_count === 1 ? "" : "s"} from{" "}
              <strong>{result.run.endpoint_count}</strong> endpoint
              {result.run.endpoint_count === 1 ? "" : "s"}
              {result.run.spec_filename ? ` in ${result.run.spec_filename}` : ""}. All cases are
              pending review.
            </p>
            <div className="flex flex-wrap gap-2 pt-2">
              <LinkButton href="/dashboard/test-cases" variant="secondary" size="sm">
                Review in Test Cases
              </LinkButton>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setResult(null);
                  setSpecText("");
                  setSpecFilename("");
                }}
              >
                Generate another
              </Button>
            </div>
          </Card>

          <Card className="p-0">
            <ul className="divide-y divide-slate-200">
              {result.cases.map((c) => (
                <li key={c.id} className="flex items-center justify-between gap-3 px-6 py-3">
                  <Link
                    href={`/dashboard/test-cases/${c.id}`}
                    className="min-w-0 flex-1 truncate text-sm font-medium text-slate-900 hover:text-brand-600"
                  >
                    {c.title}
                  </Link>
                  <div className="flex flex-none items-center gap-2">
                    <Badge variant={priorityVariants[c.priority] || "neutral"}>{c.priority}</Badge>
                    <Badge variant="warning">needs review</Badge>
                  </div>
                </li>
              ))}
            </ul>
          </Card>
        </>
      )}
    </div>
  );
}
