"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { Plus, ChevronRight, Bot } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { CardSkeleton } from "@/components/ui/skeleton";
import { Input } from "@/components/ui/input";
import { listAutomationProjects, createAutomationProject } from "@/features/automationhub/api";
import { listProjects } from "@/features/platform/api";
import type { AutomationProject, PaginationMeta } from "@/types/automationhub";
import type { Project } from "@/types/platform";

export default function AutomationProjectsPage() {
  const params = useParams();
  const [workspaceId, setWorkspaceId] = useState<string>("");

  useEffect(() => {
    if (typeof window === "undefined") return;
    const id =
      (params.workspace as string) || localStorage.getItem("testra_workspace_id") || "";
    if (id) {
      setWorkspaceId(id);
      localStorage.setItem("testra_workspace_id", id);
    }
  }, [params.workspace]);

  const [projects, setProjects] = useState<AutomationProject[]>([]);
  const [platformProjects, setPlatformProjects] = useState<Project[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [cursor, setCursor] = useState<string | undefined>(undefined);
  const [error, setError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);

  const [name, setName] = useState("");
  const [framework, setFramework] = useState("junit");
  const [projectId, setProjectId] = useState("");
  const [repo, setRepo] = useState("");
  const [branch, setBranch] = useState("");
  const [command, setCommand] = useState("");
  const [creating, setCreating] = useState(false);

  const fetchProjects = useCallback(
    async (reset?: boolean) => {
      if (!workspaceId) return;
      setLoading(true);
      setError(null);
      try {
        const result = await listAutomationProjects(workspaceId, {
          cursor: reset ? undefined : cursor,
        });
        setProjects((prev) => (reset ? result.data : [...prev, ...result.data]));
        setMeta(result.meta);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load projects");
      } finally {
        setLoading(false);
      }
    },
    [workspaceId, cursor],
  );

  useEffect(() => {
    if (typeof window !== "undefined" && workspaceId) {
      localStorage.setItem("testra_workspace_id", workspaceId);
    }
    if (workspaceId) {
      fetchProjects(true);
      listProjects(workspaceId).then(setPlatformProjects).catch(() => setPlatformProjects([]));
    } else {
      setLoading(false);
    }
  }, [workspaceId, fetchProjects]);

  useEffect(() => {
    if (cursor) fetchProjects(false);
  }, [cursor, fetchProjects]);

  const loadMore = () => {
    if (meta?.next_cursor) setCursor(meta.next_cursor);
  };

  const resetForm = () => {
    setName("");
    setRepo("");
    setBranch("");
    setCommand("");
    setProjectId("");
    setFramework("junit");
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !workspaceId) return;
    setCreating(true);
    try {
      const created = await createAutomationProject({
        workspace_id: workspaceId,
        project_id: projectId || undefined,
        name,
        framework,
        repository_url: repo,
        branch,
        command,
      });
      setProjects((prev) => [created, ...prev]);
      setShowForm(false);
      resetForm();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create project");
    } finally {
      setCreating(false);
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Automation Hub"
        description="Manage automated test projects and report imports."
        actions={
          <Button onClick={() => setShowForm((v) => !v)}>
            <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
            {showForm ? "Cancel" : "New Project"}
          </Button>
        }
      />

      {showForm && (
        <Card className="p-4">
          <h3 className="mb-4 text-sm font-medium text-slate-900">Create automation project</h3>
          <form onSubmit={handleCreate} className="space-y-4">
            <Input
              placeholder="Project name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
            <select
              className="w-full rounded-md border border-slate-300 p-2 text-sm"
              value={framework}
              onChange={(e) => setFramework(e.target.value)}
            >
              <option value="junit">JUnit XML</option>
              <option value="pytest-junit">Pytest JUnit XML</option>
              <option value="playwright">Playwright</option>
              <option value="cypress">Cypress</option>
              <option value="newman">Newman</option>
              <option value="robot">Robot Framework</option>
            </select>
            <select
              className="w-full rounded-md border border-slate-300 p-2 text-sm"
              value={projectId}
              onChange={(e) => setProjectId(e.target.value)}
            >
              <option value="">Link to project (optional)</option>
              {platformProjects.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
            <Input placeholder="Repository URL" value={repo} onChange={(e) => setRepo(e.target.value)} />
            <Input placeholder="Branch" value={branch} onChange={(e) => setBranch(e.target.value)} />
            <Input
              placeholder="Command (e.g. npx playwright test)"
              value={command}
              onChange={(e) => setCommand(e.target.value)}
            />
            <div className="flex justify-end gap-2">
              <Button type="button" variant="secondary" onClick={() => setShowForm(false)}>
                Cancel
              </Button>
              <Button type="submit" loading={creating}>
                Create
              </Button>
            </div>
          </form>
        </Card>
      )}

      {error && (
        <div role="alert">
          <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>
        </div>
      )}

      {!workspaceId ? (
        <EmptyState
          icon={Bot}
          title="No workspace selected"
          description="Select a workspace to manage automation projects."
        />
      ) : loading && projects.length === 0 ? (
        <CardSkeleton count={3} />
      ) : projects.length === 0 ? (
        <EmptyState
          icon={Bot}
          title="No automation projects yet"
          description="Create a project to start importing test reports."
        />
      ) : (
        <div className="space-y-3">
          {projects.map((project) => (
            <Link
              key={project.id}
              href={`/dashboard/automation/${project.id}`}
              className="group block"
            >
              <Card className="p-4 transition-shadow group-hover:shadow-md">
                <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                  <div>
                    <h3 className="font-medium text-slate-900">{project.name}</h3>
                    <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-slate-500">
                      <span className="rounded bg-slate-100 px-2 py-0.5">{project.framework}</span>
                      {project.repository_url && (
                        <span className="truncate max-w-xs">{project.repository_url}</span>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
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
