"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { ChevronRight, Upload, PlayCircle, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { CardSkeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  getAutomationProject,
  listAutomationExecutions,
  importAutomationExecution,
  updateAutomationProject,
} from "@/features/automationhub/api";
import type { AutomationExecution, PaginationMeta } from "@/types/automationhub";

const statusVariants: Record<string, "neutral" | "info" | "success" | "danger" | "warning"> = {
  pending: "neutral",
  running: "info",
  passed: "success",
  failed: "danger",
  skipped: "warning",
  completed: "success",
};

export default function AutomationProjectDetailPage() {
  const params = useParams();
  const projectId = (params.projectId as string) || "";

  const [project, setProject] = useState<{ name: string; framework: string; repository_url: string; branch: string; command: string } | null>(null);
  const [executions, setExecutions] = useState<AutomationExecution[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [cursor, setCursor] = useState<string | undefined>(undefined);
  const [error, setError] = useState<string | null>(null);

  const [showImport, setShowImport] = useState(false);
  const [reportName, setReportName] = useState("");
  const [format, setFormat] = useState("junit");
  const [file, setFile] = useState<File | null>(null);
  const [importing, setImporting] = useState(false);
  const [autoDefects, setAutoDefects] = useState(true);
  const [mapCases, setMapCases] = useState(false);

  const fetchExecutions = useCallback(
    async (reset?: boolean) => {
      if (!projectId) return;
      setLoading(true);
      setError(null);
      try {
        const result = await listAutomationExecutions(projectId, {
          cursor: reset ? undefined : cursor,
        });
        setExecutions((prev) => (reset ? result.data : [...prev, ...result.data]));
        setMeta(result.meta);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load executions");
      } finally {
        setLoading(false);
      }
    },
    [projectId, cursor],
  );

  useEffect(() => {
    if (!projectId) return;
    getAutomationProject(projectId)
      .then((p) =>
        setProject({
          name: p.name,
          framework: p.framework,
          repository_url: p.repository_url,
          branch: p.branch,
          command: p.command,
        }),
      )
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load project"));
    fetchExecutions(true);
  }, [projectId, fetchExecutions]);

  useEffect(() => {
    if (cursor) fetchExecutions(false);
  }, [cursor, fetchExecutions]);

  const loadMore = () => {
    if (meta?.next_cursor) setCursor(meta.next_cursor);
  };

  const handleImport = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file || !projectId) return;
    setImporting(true);
    try {
      await importAutomationExecution(projectId, {
        name: reportName || file.name,
        format,
        report: file,
        auto_create_defects: autoDefects,
        map_test_cases: mapCases,
      });
      setShowImport(false);
      setFile(null);
      setReportName("");
      fetchExecutions(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to import report");
    } finally {
      setImporting(false);
    }
  };

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!project || !projectId) return;
    try {
      const updated = await updateAutomationProject(projectId, project);
      setProject({
        name: updated.name,
        framework: updated.framework,
        repository_url: updated.repository_url,
        branch: updated.branch,
        command: updated.command,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update project");
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2 text-sm text-slate-500">
        <Link href="/dashboard/automation" className="flex items-center hover:text-slate-900">
          <ArrowLeft className="mr-1 h-4 w-4" />
          Automation Hub
        </Link>
      </div>

      <PageHeader
        title={project?.name || "Project"}
        description="View executions, import reports, and manage settings."
        actions={
          <Button onClick={() => setShowImport((v) => !v)}>
            <Upload className="mr-2 h-4 w-4" aria-hidden="true" />
            Import Report
          </Button>
        }
      />

      {error && (
        <div role="alert">
          <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>
        </div>
      )}

      {project && (
        <Card className="p-4">
          <h3 className="mb-4 text-sm font-medium text-slate-900">Settings</h3>
          <form onSubmit={handleUpdate} className="space-y-4">
            <Input
              value={project.name}
              onChange={(e) => setProject({ ...project, name: e.target.value })}
              placeholder="Name"
            />
            <select
              className="w-full rounded-md border border-slate-300 p-2 text-sm"
              value={project.framework}
              onChange={(e) => setProject({ ...project, framework: e.target.value })}
            >
              <option value="junit">JUnit XML</option>
              <option value="pytest-junit">Pytest JUnit XML</option>
              <option value="playwright">Playwright</option>
              <option value="cypress">Cypress</option>
              <option value="newman">Newman</option>
              <option value="robot">Robot Framework</option>
            </select>
            <Input
              value={project.repository_url}
              onChange={(e) => setProject({ ...project, repository_url: e.target.value })}
              placeholder="Repository URL"
            />
            <Input
              value={project.branch}
              onChange={(e) => setProject({ ...project, branch: e.target.value })}
              placeholder="Branch"
            />
            <Input
              value={project.command}
              onChange={(e) => setProject({ ...project, command: e.target.value })}
              placeholder="Command"
            />
            <div className="flex justify-end">
              <Button type="submit">Save Settings</Button>
            </div>
          </form>
        </Card>
      )}

      {showImport && (
        <Card className="p-4">
          <h3 className="mb-4 text-sm font-medium text-slate-900">Import report</h3>
          <form onSubmit={handleImport} className="space-y-4">
            <Input
              placeholder="Report name"
              value={reportName}
              onChange={(e) => setReportName(e.target.value)}
            />
            <select
              className="w-full rounded-md border border-slate-300 p-2 text-sm"
              value={format}
              onChange={(e) => setFormat(e.target.value)}
            >
              <option value="junit">JUnit XML</option>
              <option value="pytest-junit">Pytest JUnit XML</option>
              <option value="playwright">Playwright</option>
              <option value="cypress">Cypress</option>
              <option value="newman">Newman</option>
              <option value="robot">Robot Framework</option>
            </select>
            <Input
              type="file"
              accept=".xml,.json,.html,.txt"
              onChange={(e) => setFile(e.target.files?.[0] || null)}
              required
            />
            <div className="flex items-center gap-4 text-sm">
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={autoDefects}
                  onChange={(e) => setAutoDefects(e.target.checked)}
                />
                Auto-create defects
              </label>
              <label className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={mapCases}
                  onChange={(e) => setMapCases(e.target.checked)}
                />
                Map test cases
              </label>
            </div>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="secondary" onClick={() => setShowImport(false)}>
                Cancel
              </Button>
              <Button type="submit" loading={importing}>
                Import
              </Button>
            </div>
          </form>
        </Card>
      )}

      <h3 className="text-sm font-medium text-slate-900">Execution History</h3>

      {loading && executions.length === 0 ? (
        <CardSkeleton count={3} />
      ) : executions.length === 0 ? (
        <EmptyState
          icon={PlayCircle}
          title="No executions yet"
          description="Import a report to create your first execution."
        />
      ) : (
        <div className="space-y-3">
          {executions.map((exec) => (
            <Link
              key={exec.id}
              href={`/dashboard/automation/${projectId}/executions/${exec.id}`}
              className="group block"
            >
              <Card className="p-4 transition-shadow group-hover:shadow-md">
                <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <h3 className="font-medium text-slate-900">{exec.name}</h3>
                    <div className="mt-1 flex flex-wrap items-center gap-2 text-xs">
                      <Badge variant={statusVariants[exec.status] || "neutral"}>{exec.status}</Badge>
                      <span className="text-slate-500">
                        {exec.passed} passed · {exec.failed} failed · {exec.skipped} skipped · {exec.total} total
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-xs text-slate-500">
                      {new Date(exec.created_at).toLocaleString()}
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
