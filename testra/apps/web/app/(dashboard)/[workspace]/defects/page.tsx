"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { Bug, Plus, ChevronRight, Filter } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { CardSkeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { listDefects, createDefect } from "@/features/defects/api";
import type { Defect, PaginationMeta } from "@/types/defects";

const statusVariants: Record<string, "neutral" | "info" | "success" | "danger" | "warning"> = {
  open: "info",
  in_progress: "warning",
  resolved: "success",
  closed: "neutral",
  rejected: "danger",
};

const severityVariants: Record<string, "default" | "info" | "warning" | "danger"> = {
  low: "default",
  medium: "info",
  high: "warning",
  critical: "danger",
};

export default function DefectsPage() {
  const [defects, setDefects] = useState<Defect[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [cursor, setCursor] = useState<string | undefined>(undefined);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [creating, setCreating] = useState(false);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [severity, setSeverity] = useState<Defect["severity"]>("medium");
  const [priority, setPriority] = useState<Defect["priority"]>("medium");

  const projectId =
    typeof window !== "undefined"
      ? localStorage.getItem("testra_project_id") || ""
      : "";

  const fetchDefects = useCallback(
    async (reset?: boolean) => {
      setLoading(true);
      setError(null);
      try {
        const result = await listDefects(projectId, {
          cursor: reset ? undefined : cursor,
        });
        setDefects((prev) => (reset ? result.data : [...prev, ...result.data]));
        setMeta(result.meta);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load defects");
      } finally {
        setLoading(false);
      }
    },
    [projectId, cursor],
  );

  useEffect(() => {
    if (projectId) {
      fetchDefects(true);
    } else {
      setLoading(false);
    }
  }, [projectId, fetchDefects]);

  useEffect(() => {
    if (cursor) {
      fetchDefects(false);
    }
  }, [cursor, fetchDefects]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!projectId) return;
    const workspaceId = localStorage.getItem("testra_workspace_id") || "";
    if (!workspaceId) return;

    setCreating(true);
    setError(null);
    try {
      const defect = await createDefect({
        workspace_id: workspaceId,
        project_id: projectId,
        title: title.trim(),
        description: description.trim(),
        severity,
        priority,
      });
      setDefects((prev) => [defect, ...prev]);
      setTitle("");
      setDescription("");
      setSeverity("medium");
      setPriority("medium");
      setShowCreate(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create defect");
    } finally {
      setCreating(false);
    }
  }

  const loadMore = () => {
    if (meta?.next_cursor) {
      setCursor(meta.next_cursor);
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Defects"
        description="Track bugs and link failures to test runs."
        actions={
          <Button onClick={() => setShowCreate((s) => !s)}>
            <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
            {showCreate ? "Cancel" : "New Defect"}
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

      {showCreate && (
        <Card className="p-6">
          <form onSubmit={handleCreate} className="space-y-4">
            <Input
              label="Title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="e.g. Login button unresponsive"
              required
            />
            <Input
              label="Description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Steps to reproduce..."
            />
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block space-y-1">
                <span className="text-sm font-medium text-slate-700">Severity</span>
                <select
                  value={severity}
                  onChange={(e) => setSeverity(e.target.value as Defect["severity"])}
                  className="w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
                >
                  <option value="low">Low</option>
                  <option value="medium">Medium</option>
                  <option value="high">High</option>
                  <option value="critical">Critical</option>
                </select>
              </label>
              <label className="block space-y-1">
                <span className="text-sm font-medium text-slate-700">Priority</span>
                <select
                  value={priority}
                  onChange={(e) => setPriority(e.target.value as Defect["priority"])}
                  className="w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
                >
                  <option value="low">Low</option>
                  <option value="medium">Medium</option>
                  <option value="high">High</option>
                  <option value="critical">Critical</option>
                </select>
              </label>
            </div>
            <div className="flex gap-2">
              <Button type="submit" loading={creating}>
                Create defect
              </Button>
              <Button variant="secondary" onClick={() => setShowCreate(false)}>
                Cancel
              </Button>
            </div>
          </form>
        </Card>
      )}

      {!projectId ? (
        <EmptyState
          icon={Filter}
          title="No project selected"
          description="Select a project from the Projects page to view and create defects."
          action={{ label: "Go to Projects", href: "/dashboard/projects" }}
        />
      ) : loading && defects.length === 0 ? (
        <CardSkeleton count={3} />
      ) : defects.length === 0 ? (
        <EmptyState
          icon={Bug}
          title="No defects yet"
          description="Create your first defect to track a bug or failed test."
          action={{ label: "New Defect", onClick: () => setShowCreate(true) }}
        />
      ) : (
        <div className="space-y-3">
          {defects.map((defect) => (
            <Link key={defect.id} href={`/dashboard/defects/${defect.id}`} className="group block">
              <Card className="p-4 transition-shadow group-hover:shadow-md">
                <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <h3 className="font-medium text-slate-900">{defect.title}</h3>
                    <div className="mt-1 flex flex-wrap items-center gap-2">
                      <Badge variant={statusVariants[defect.status] || "neutral"}>
                        {defect.status.replace("_", " ")}
                      </Badge>
                      <Badge variant={severityVariants[defect.severity] || "default"}>
                        {defect.severity}
                      </Badge>
                      <Badge variant="neutral">{defect.priority}</Badge>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-xs text-slate-500">
                      {new Date(defect.updated_at).toLocaleString()}
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

