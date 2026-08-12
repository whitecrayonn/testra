"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { ArrowLeft, Paperclip, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { CardSkeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  getAutomationExecution,
  listAutomationArtifacts,
  uploadAutomationArtifact,
  getAutomationArtifactDownloadUrl,
  listAutomationLogs,
} from "@/features/automationhub/api";
import type { AutomationArtifact, AutomationLog, PaginationMeta } from "@/types/automationhub";

const statusVariants: Record<string, "neutral" | "info" | "success" | "danger" | "warning"> = {
  pending: "neutral",
  running: "info",
  passed: "success",
  failed: "danger",
  skipped: "warning",
  completed: "success",
};

const workspaceId = () => {
  if (typeof window === "undefined") return "";
  return localStorage.getItem("testra_workspace_id") || "";
};

export default function AutomationExecutionDetailPage() {
  const params = useParams();
  const projectId = (params.projectId as string) || "";
  const executionId = (params.executionId as string) || "";

  const [exec, setExec] = useState<{
    name: string;
    status: string;
    total: number;
    passed: number;
    failed: number;
    skipped: number;
    blocked: number;
    duration_ms: number;
    report_format: string;
    created_at: string;
  } | null>(null);
  const [artifacts, setArtifacts] = useState<AutomationArtifact[]>([]);
  const [artifactMeta, setArtifactMeta] = useState<PaginationMeta | null>(null);
  const [logs, setLogs] = useState<AutomationLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [error, setError] = useState<string | null>(null);

  const fetchAll = useCallback(async () => {
    if (!executionId) return;
    setLoading(true);
    setError(null);
    try {
      const [e, a, l] = await Promise.all([
        getAutomationExecution(executionId),
        listAutomationArtifacts(executionId),
        listAutomationLogs(executionId),
      ]);
      setExec({
        name: e.name,
        status: e.status,
        total: e.total,
        passed: e.passed,
        failed: e.failed,
        skipped: e.skipped,
        blocked: e.blocked,
        duration_ms: e.duration_ms,
        report_format: e.report_format,
        created_at: e.created_at,
      });
      setArtifacts(a.data);
      setArtifactMeta(a.meta);
      setLogs(l.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load execution");
    } finally {
      setLoading(false);
    }
  }, [executionId]);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file || !executionId) return;
    setUploading(true);
    try {
      await uploadAutomationArtifact(executionId, {
        workspace_id: workspaceId(),
        file,
      });
      setFile(null);
      const a = await listAutomationArtifacts(executionId);
      setArtifacts(a.data);
      setArtifactMeta(a.meta);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to upload artifact");
    } finally {
      setUploading(false);
    }
  };

  if (loading && !exec) return <CardSkeleton count={3} />;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2 text-sm text-slate-500">
        <Link
          href={`/dashboard/automation/${projectId}`}
          className="flex items-center hover:text-slate-900"
        >
          <ArrowLeft className="mr-1 h-4 w-4" />
          Back to project
        </Link>
      </div>

      <PageHeader
        title={exec?.name || "Execution"}
        description={`Status and artifacts for this automation run.`}
      />

      {error && (
        <div role="alert">
          <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>
        </div>
      )}

      {exec && (
        <Card className="p-4">
          <div className="flex flex-wrap items-center gap-4">
            <Badge variant={statusVariants[exec.status] || "neutral"}>{exec.status}</Badge>
            <span className="text-sm text-slate-600">
              {exec.passed} passed · {exec.failed} failed · {exec.skipped} skipped · {exec.blocked} blocked · {exec.total} total
            </span>
            <span className="text-sm text-slate-500">
              {exec.duration_ms}ms · {exec.report_format}
            </span>
            <span className="ml-auto text-xs text-slate-400">
              {new Date(exec.created_at).toLocaleString()}
            </span>
          </div>
        </Card>
      )}

      <div className="grid gap-6 lg:grid-cols-2">
        <Card className="p-4">
          <h3 className="mb-4 text-sm font-medium text-slate-900 flex items-center gap-2">
            <Paperclip className="h-4 w-4" />
            Artifacts
          </h3>
          <form onSubmit={handleUpload} className="mb-4 flex gap-2">
            <Input
              type="file"
              onChange={(e) => setFile(e.target.files?.[0] || null)}
              className="flex-1"
            />
            <Button type="submit" loading={uploading}>
              Upload
            </Button>
          </form>
          {artifacts.length === 0 ? (
            <EmptyState icon={FileText} title="No artifacts" description="Upload files for this execution." />
          ) : (
            <ul className="space-y-2">
              {artifacts.map((a) => (
                <li key={a.id} className="flex items-center justify-between rounded border border-slate-200 p-2 text-sm">
                  <span className="truncate font-medium text-slate-700">{a.name}</span>
                  <a
                    href={getAutomationArtifactDownloadUrl(a.id)}
                    download={a.name}
                    className="text-brand-600 hover:underline"
                  >
                    Download
                  </a>
                </li>
              ))}
              {artifactMeta?.has_more && (
                <li className="text-center text-sm text-slate-500">More artifacts available</li>
              )}
            </ul>
          )}
        </Card>

        <Card className="p-4">
          <h3 className="mb-4 text-sm font-medium text-slate-900 flex items-center gap-2">
            <FileText className="h-4 w-4" />
            Logs
          </h3>
          {logs.length === 0 ? (
            <EmptyState icon={FileText} title="No logs" description="Execution logs will appear here." />
          ) : (
            <ul className="max-h-96 space-y-2 overflow-auto">
              {logs.map((l) => (
                <li key={l.id} className="rounded border border-slate-200 p-2 text-sm">
                  <span className="rounded bg-slate-100 px-1.5 py-0.5 text-xs font-medium">{l.level}</span>
                  <p className="mt-1 text-slate-700">{l.message}</p>
                  <p className="text-xs text-slate-400">{new Date(l.logged_at).toLocaleString()}</p>
                </li>
              ))}
            </ul>
          )}
        </Card>
      </div>
    </div>
  );
}
