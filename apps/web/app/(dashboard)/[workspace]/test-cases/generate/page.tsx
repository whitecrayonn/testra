"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { ArrowLeft, Upload, FileJson, FileSpreadsheet, CheckCircle2, Zap, Play } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { LinkButton } from "@/components/ui/link-button";
import { ResponseViewer } from "@/components/apitesting/response-viewer";
import { generateFromSpec, generateFromEndpoint, generateFromFile } from "@/features/testgen/api";
import { executeRequest } from "@/features/apitesting/api";
import type {
  GenerateFromSpecResponse,
  GenerateFromFileResponse,
  EndpointField,
  EndpointFieldType,
} from "@/types/testgen";
import { HTTP_METHODS } from "@/types/apitesting";
import type { HTTPMethod, ExecutionResult } from "@/types/apitesting";

const priorityVariants: Record<string, "neutral" | "info" | "warning" | "danger"> = {
  low: "neutral",
  medium: "info",
  high: "warning",
  critical: "danger",
};

const FIELD_TYPES: EndpointFieldType[] = ["string", "integer", "number", "boolean"];

const MAX_UPLOAD_BYTES = 5 * 1024 * 1024; // mirrors the backend's route-scoped limit

interface FieldRow {
  name: string;
  value: string;
  type: EndpointFieldType;
  required: boolean;
}

function emptyFieldRow(): FieldRow {
  return { name: "", value: "", type: "string", required: false };
}

// The happy-path case is always titled "<METHOD> <path> — valid request
// succeeds" (see apps/api/internal/testgen/openapi_parser.go) — it's the
// only generated case with concrete example values, so it's the only one
// "Test Now" can execute directly.
function isHappyPathCase(title: string) {
  return title.includes("valid request succeeds");
}

function coerceFieldValue(row: FieldRow): unknown {
  if (row.type === "integer" || row.type === "number") {
    const n = Number(row.value);
    return Number.isNaN(n) ? row.value : n;
  }
  if (row.type === "boolean") {
    return row.value === "true";
  }
  return row.value;
}

function FieldRowsEditor({
  rows,
  onChange,
  showType = true,
}: {
  rows: FieldRow[];
  onChange: (rows: FieldRow[]) => void;
  showType?: boolean;
}) {
  function update(index: number, patch: Partial<FieldRow>) {
    const next = [...rows];
    next[index] = { ...next[index], ...patch };
    onChange(next);
  }
  function add() {
    onChange([...rows, emptyFieldRow()]);
  }
  function remove(index: number) {
    const next = [...rows];
    next.splice(index, 1);
    onChange(next);
  }

  return (
    <div className="space-y-2">
      <table className="w-full text-sm">
        <thead className="text-left text-slate-500">
          <tr>
            <th className="pb-2 font-medium">Name</th>
            <th className="pb-2 font-medium">Example value</th>
            {showType && <th className="pb-2 font-medium w-28">Type</th>}
            <th className="pb-2 font-medium w-20">Required</th>
            <th className="pb-2 font-medium w-10"></th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} className="border-b border-slate-100">
              <td className="py-2 pr-2">
                <Input value={row.name} onChange={(e) => update(i, { name: e.target.value })} className="h-8" />
              </td>
              <td className="py-2 pr-2">
                <Input value={row.value} onChange={(e) => update(i, { value: e.target.value })} className="h-8" />
              </td>
              {showType && (
                <td className="py-2 pr-2">
                  <select
                    value={row.type}
                    onChange={(e) => update(i, { type: e.target.value as EndpointFieldType })}
                    className="h-8 w-full rounded-md border border-slate-300 bg-white px-2 text-sm text-slate-900"
                  >
                    {FIELD_TYPES.map((t) => (
                      <option key={t} value={t}>
                        {t}
                      </option>
                    ))}
                  </select>
                </td>
              )}
              <td className="py-2 pr-2">
                <Switch checked={row.required} onCheckedChange={(checked) => update(i, { required: checked })} />
              </td>
              <td className="py-2">
                <button
                  type="button"
                  onClick={() => remove(i)}
                  className="text-xs text-slate-400 hover:text-red-600"
                  aria-label="Remove"
                >
                  Remove
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <Button type="button" variant="secondary" size="sm" onClick={add}>
        Add
      </Button>
    </div>
  );
}

export default function GenerateTestCasesPage() {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [projectId, setProjectId] = useState("");
  const [workspaceId, setWorkspaceId] = useState("");

  const [mode, setMode] = useState<"endpoint" | "spec" | "upload">("endpoint");

  // Full OpenAPI spec upload — unchanged from the original flow.
  const [specText, setSpecText] = useState("");
  const [specFilename, setSpecFilename] = useState("");

  // Single-endpoint quick generate.
  const [method, setMethod] = useState<HTTPMethod>("GET");
  const [path, setPath] = useState("");
  const [requiresAuth, setRequiresAuth] = useState(false);
  const [pathFields, setPathFields] = useState<FieldRow[]>([]);
  const [queryFields, setQueryFields] = useState<FieldRow[]>([]);
  const [headerFields, setHeaderFields] = useState<FieldRow[]>([]);
  const [bodyFields, setBodyFields] = useState<FieldRow[]>([]);

  // Upload a spreadsheet — AI-assisted (see handleGenerateFromFile).
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [uploadContext, setUploadContext] = useState("");

  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<GenerateFromSpecResponse | GenerateFromFileResponse | null>(null);

  const [testingNow, setTestingNow] = useState(false);
  const [testNowResult, setTestNowResult] = useState<ExecutionResult | null>(null);
  const [testNowError, setTestNowError] = useState<string | null>(null);

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

  function resetResult() {
    setResult(null);
    setTestNowResult(null);
    setTestNowError(null);
  }

  async function handleGenerateFromSpec() {
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
    resetResult();
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

  async function handleGenerateFromEndpoint() {
    if (!projectId || !workspaceId) {
      setError("No project or workspace selected. Select a project first.");
      return;
    }
    if (!path.trim()) {
      setError("Enter an endpoint path or URL first.");
      return;
    }

    const toFields = (rows: FieldRow[], location: EndpointField["location"]): EndpointField[] =>
      rows
        .filter((r) => r.name.trim() !== "")
        .map((r) => ({ name: r.name.trim(), location, type: r.type, required: r.required }));

    const fields: EndpointField[] = [
      ...toFields(pathFields, "path"),
      ...toFields(queryFields, "query"),
      ...toFields(headerFields, "header"),
      ...toFields(bodyFields, "body"),
    ];

    setGenerating(true);
    setError(null);
    resetResult();
    try {
      const response = await generateFromEndpoint({
        workspace_id: workspaceId,
        project_id: projectId,
        method,
        path: path.trim(),
        fields,
        requires_auth: requiresAuth,
      });
      setResult(response);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to generate test cases");
    } finally {
      setGenerating(false);
    }
  }

  async function handleGenerateFromFile() {
    if (!projectId || !workspaceId) {
      setError("No project or workspace selected. Select a project first.");
      return;
    }
    if (!uploadFile) {
      setError("Choose a .csv or .xlsx file first.");
      return;
    }
    const name = uploadFile.name.toLowerCase();
    if (!name.endsWith(".csv") && !name.endsWith(".xlsx")) {
      setError("Only .csv or .xlsx files are supported.");
      return;
    }
    if (uploadFile.size > MAX_UPLOAD_BYTES) {
      setError("File is too large — max 5 MB.");
      return;
    }

    setGenerating(true);
    setError(null);
    resetResult();
    try {
      const response = await generateFromFile({
        workspace_id: workspaceId,
        project_id: projectId,
        file: uploadFile,
        context: uploadContext.trim() || undefined,
      });
      setResult(response);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to generate test cases");
    } finally {
      setGenerating(false);
    }
  }

  async function handleTestNow(testCaseId: string) {
    setTestingNow(true);
    setTestNowError(null);
    setTestNowResult(null);
    try {
      let url = path.trim();
      for (const row of pathFields) {
        if (row.name.trim()) {
          url = url.replace(`{${row.name.trim()}}`, encodeURIComponent(row.value));
        }
      }

      const bodyRows = bodyFields.filter((r) => r.name.trim() !== "");
      const bodyContent =
        bodyRows.length > 0
          ? JSON.stringify(Object.fromEntries(bodyRows.map((r) => [r.name.trim(), coerceFieldValue(r)])))
          : "";

      const response = await executeRequest({
        workspace_id: workspaceId,
        test_case_id: testCaseId,
        save: true,
        request: {
          method,
          url,
          headers: headerFields
            .filter((r) => r.name.trim() !== "")
            .map((r) => ({ key: r.name.trim(), value: r.value, enabled: true })),
          query_params: queryFields
            .filter((r) => r.name.trim() !== "")
            .map((r) => ({ key: r.name.trim(), value: r.value, enabled: true })),
          body_type: bodyContent ? "json" : "none",
          body_content: bodyContent,
        },
      });
      setTestNowResult(response.result);
    } catch (err) {
      setTestNowError(err instanceof Error ? err.message : "Failed to execute the request");
    } finally {
      setTestingNow(false);
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

  const happyPathCase = result?.cases.find((c) => isHappyPathCase(c.title));

  return (
    <div className="space-y-6">
      <PageHeader
        title="Generate Test Cases"
        description="From an OpenAPI spec: deterministic, rule-based, no AI. From an uploaded spreadsheet: AI-assisted."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Test Cases", href: "/dashboard/test-cases" },
          { label: "Generate" },
        ]}
        actions={
          !result ? (
            <Button
              onClick={
                mode === "endpoint"
                  ? handleGenerateFromEndpoint
                  : mode === "spec"
                    ? handleGenerateFromSpec
                    : handleGenerateFromFile
              }
              loading={generating}
            >
              <FileJson className="mr-2 h-4 w-4" aria-hidden="true" />
              Generate
            </Button>
          ) : undefined
        }
      />

      {error && (
        <div role="alert">
          <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>
        </div>
      )}

      {!result && (
        <>
          <div className="flex gap-2">
            <Button
              type="button"
              variant={mode === "endpoint" ? "primary" : "secondary"}
              size="sm"
              onClick={() => setMode("endpoint")}
            >
              <Zap className="mr-2 h-4 w-4" aria-hidden="true" />
              Single Endpoint
            </Button>
            <Button
              type="button"
              variant={mode === "spec" ? "primary" : "secondary"}
              size="sm"
              onClick={() => setMode("spec")}
            >
              <Upload className="mr-2 h-4 w-4" aria-hidden="true" />
              Full OpenAPI Spec
            </Button>
            <Button
              type="button"
              variant={mode === "upload" ? "primary" : "secondary"}
              size="sm"
              onClick={() => setMode("upload")}
            >
              <FileSpreadsheet className="mr-2 h-4 w-4" aria-hidden="true" />
              Upload Test Cases
            </Button>
          </div>

          {mode === "endpoint" ? (
            <Card className="space-y-5 p-6">
              <p className="text-sm text-slate-600">
                Describe one endpoint — method, path, and its parameters — and get draft test
                cases without writing an OpenAPI spec. Every case is still{" "}
                <strong>pending_review</strong> and must be approved before it counts toward
                coverage.
              </p>

              <div className="flex flex-wrap items-center gap-3">
                <select
                  value={method}
                  onChange={(e) => setMethod(e.target.value as HTTPMethod)}
                  className="h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm font-semibold text-slate-900 focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
                >
                  {HTTP_METHODS.map((m) => (
                    <option key={m} value={m}>
                      {m}
                    </option>
                  ))}
                </select>
                <Input
                  value={path}
                  onChange={(e) => setPath(e.target.value)}
                  placeholder="/users/{id} or https://api.example.com/users/{id}"
                  className="min-w-[20rem] flex-1"
                />
                <label className="flex items-center gap-2 text-sm text-slate-700">
                  <Switch checked={requiresAuth} onCheckedChange={setRequiresAuth} />
                  Requires authentication
                </label>
              </div>

              <div>
                <h3 className="mb-2 text-sm font-medium text-slate-700">Path parameters</h3>
                <FieldRowsEditor rows={pathFields} onChange={setPathFields} />
              </div>
              <div>
                <h3 className="mb-2 text-sm font-medium text-slate-700">Query parameters</h3>
                <FieldRowsEditor rows={queryFields} onChange={setQueryFields} />
              </div>
              <div>
                <h3 className="mb-2 text-sm font-medium text-slate-700">Headers</h3>
                <FieldRowsEditor rows={headerFields} onChange={setHeaderFields} />
              </div>
              <div>
                <h3 className="mb-2 text-sm font-medium text-slate-700">Body fields (JSON)</h3>
                <FieldRowsEditor rows={bodyFields} onChange={setBodyFields} />
              </div>
            </Card>
          ) : mode === "spec" ? (
            <Card className="space-y-4 p-6">
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
                  JSON only (not YAML) — this keeps the generator dependency-free. Convert YAML to
                  JSON first if needed.
                </p>
              </div>
            </Card>
          ) : (
            <Card className="space-y-4 p-6">
              <p className="text-sm text-slate-600">
                Upload a spreadsheet of test cases or raw requirements. AI reads each row and
                drafts a standard test case (title, steps, expected result, priority) from it —
                works best with columns like <strong>Title</strong>, <strong>Steps</strong>,{" "}
                <strong>Expected Result</strong>, and <strong>Priority</strong>, but other layouts
                are still attempted. Every case is still <strong>pending_review</strong>, and rows
                the AI can&apos;t confidently map are reported below instead of guessed at.
              </p>

              <div>
                <label htmlFor="upload-file" className="mb-1 block text-sm font-medium text-slate-700">
                  File (.csv or .xlsx, max 5 MB)
                </label>
                <Input
                  id="upload-file"
                  type="file"
                  accept=".csv,.xlsx"
                  onChange={(e) => setUploadFile(e.target.files?.[0] ?? null)}
                />
                {uploadFile && <p className="mt-1 text-xs text-slate-500">{uploadFile.name}</p>}
              </div>

              <div>
                <label htmlFor="upload-context" className="mb-1 block text-sm font-medium text-slate-700">
                  Context (optional)
                </label>
                <Input
                  id="upload-context"
                  value={uploadContext}
                  onChange={(e) => setUploadContext(e.target.value)}
                  placeholder="e.g. Checkout flow, Login feature"
                />
              </div>
            </Card>
          )}
        </>
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
              <strong>{result.run.endpoint_count}</strong>{" "}
              {mode === "upload" ? "row" : "endpoint"}
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
                  resetResult();
                  setSpecText("");
                  setSpecFilename("");
                  setUploadFile(null);
                  setUploadContext("");
                }}
              >
                Generate another
              </Button>
            </div>
          </Card>

          {"skipped_rows" in result && result.skipped_rows.length > 0 && (
            <Card className="space-y-2 p-6">
              <h3 className="font-medium text-slate-900">Skipped rows</h3>
              <p className="text-sm text-slate-600">
                The AI couldn&apos;t confidently turn these rows into test cases, so nothing was
                fabricated for them:
              </p>
              <ul className="space-y-1 text-sm text-slate-600">
                {result.skipped_rows.map((s, i) => (
                  <li key={i}>
                    <span className="font-mono text-xs text-slate-400">row {s.row}</span> —{" "}
                    {s.reason}
                  </li>
                ))}
              </ul>
            </Card>
          )}

          {mode === "endpoint" && happyPathCase && (
            <Card className="space-y-3 p-6">
              <div className="flex items-center justify-between gap-3">
                <div>
                  <h3 className="font-medium text-slate-900">Test the happy path now</h3>
                  <p className="text-sm text-slate-600">
                    Sends {method} {path} using the example values above via Testra&apos;s built-in
                    API tester and links the result to this test case.
                  </p>
                </div>
                <Button
                  size="sm"
                  onClick={() => handleTestNow(happyPathCase.id)}
                  loading={testingNow}
                >
                  <Play className="mr-2 h-4 w-4" aria-hidden="true" />
                  Test Now
                </Button>
              </div>
              {testNowError && (
                <p className="text-sm text-red-600" role="alert">
                  {testNowError}
                </p>
              )}
              {testNowResult && <ResponseViewer result={testNowResult} />}
            </Card>
          )}

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

          <p className="text-sm text-slate-500">
            Testing with your own CI or tools instead? Push results via{" "}
            <Link href="/dashboard/automation" className="text-brand-600 hover:underline">
              automation hub ingestion
            </Link>{" "}
            — cases with matching titles link automatically.
          </p>
        </>
      )}
    </div>
  );
}
