"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { Plus, ChevronRight, Bot } from "lucide-react";
import { StatusPill } from "@/components/ui/status-pill";
import { listAutomationProjects, createAutomationProject } from "@/features/automationhub/api";
import { listProjects } from "@/features/platform/api";
import type { AutomationProject, PaginationMeta } from "@/types/automationhub";
import type { Project } from "@/types/platform";

export default function AutomationProjectsPage() {
  const params = useParams();
  const [workspaceId, setWorkspaceId] = useState<string>("");

  useEffect(() => {
    if (typeof window === "undefined") return;
    const id = (params.workspace as string) || localStorage.getItem("testra_workspace_id") || "";
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
    <div className="flex flex-col gap-3.5">
      <div className="flex animate-rise-sm items-end justify-between gap-4">
        <div>
          <h1 className="m-0 text-[27px] font-bold tracking-tight text-fg">Automation Hub</h1>
          <p className="mt-1.5 text-[13px] text-fg2">Manage automated test projects and report imports.</p>
        </div>
        <button
          onClick={() => setShowForm((v) => !v)}
          className="flex h-[38px] items-center gap-2 rounded-[13px] bg-gradient-to-br from-acc to-acc2 px-4 text-[13px] font-bold text-white shadow-[0_10px_26px_-12px_var(--ring)] transition-transform hover:-translate-y-px"
        >
          <Plus className="h-3.5 w-3.5" aria-hidden="true" />
          {showForm ? "Cancel" : "New Project"}
        </button>
      </div>

      {error && (
        <div role="alert" className="rounded-[16px] border border-fail-soft bg-fail-soft p-4 text-[13px] text-fail">
          {error}
        </div>
      )}

      {showForm && (
        <form
          onSubmit={handleCreate}
          className="flex animate-pop flex-col gap-4 rounded-[20px] border border-hair bg-panel p-5 shadow-glass"
        >
          <h3 className="m-0 text-[13px] font-semibold text-fg">Create automation project</h3>
          <input
            placeholder="Project name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="h-10 rounded-xl border border-hair bg-panel-hi px-3 text-[13px] text-fg placeholder-fg3 outline-none transition-colors focus:border-hair-hi"
          />
          <select
            value={framework}
            onChange={(e) => setFramework(e.target.value)}
            className="h-10 rounded-xl border border-hair bg-panel-hi px-3 text-[13px] text-fg outline-none transition-colors focus:border-hair-hi"
          >
            <option value="junit">JUnit XML</option>
            <option value="pytest-junit">Pytest JUnit XML</option>
            <option value="playwright">Playwright</option>
            <option value="cypress">Cypress</option>
            <option value="newman">Newman</option>
            <option value="robot">Robot Framework</option>
          </select>
          <select
            value={projectId}
            onChange={(e) => setProjectId(e.target.value)}
            className="h-10 rounded-xl border border-hair bg-panel-hi px-3 text-[13px] text-fg outline-none transition-colors focus:border-hair-hi"
          >
            <option value="">Link to project (optional)</option>
            {platformProjects.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
          <input
            placeholder="Repository URL"
            value={repo}
            onChange={(e) => setRepo(e.target.value)}
            className="h-10 rounded-xl border border-hair bg-panel-hi px-3 text-[13px] text-fg placeholder-fg3 outline-none transition-colors focus:border-hair-hi"
          />
          <input
            placeholder="Branch"
            value={branch}
            onChange={(e) => setBranch(e.target.value)}
            className="h-10 rounded-xl border border-hair bg-panel-hi px-3 text-[13px] text-fg placeholder-fg3 outline-none transition-colors focus:border-hair-hi"
          />
          <input
            placeholder="Command (e.g. npx playwright test)"
            value={command}
            onChange={(e) => setCommand(e.target.value)}
            className="h-10 rounded-xl border border-hair bg-panel-hi px-3 text-[13px] text-fg placeholder-fg3 outline-none transition-colors focus:border-hair-hi"
          />
          <div className="flex justify-end gap-2.5">
            <button
              type="button"
              onClick={() => setShowForm(false)}
              className="h-9 rounded-[13px] border border-hair-hi bg-panel px-4 text-[12.5px] font-semibold text-fg transition-colors hover:bg-panel-hi"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={creating}
              className="flex h-9 items-center gap-2 rounded-[13px] bg-gradient-to-br from-acc to-acc2 px-4 text-[12.5px] font-bold text-white disabled:pointer-events-none disabled:opacity-60"
            >
              {creating ? "Creating…" : "Create"}
            </button>
          </div>
        </form>
      )}

      <div className="flex flex-col gap-2">
        {!workspaceId ? (
          <EmptyPanel icon={Bot} title="No workspace selected" description="Select a workspace to manage automation projects." />
        ) : loading && projects.length === 0 ? (
          <ListSkeleton count={3} />
        ) : projects.length === 0 ? (
          <EmptyPanel icon={Bot} title="No automation projects yet" description="Create a project to start importing test reports." />
        ) : (
          <>
            {projects.map((project, i) => (
              <Link
                key={project.id}
                href={`/dashboard/automation/${project.id}`}
                style={{ animationDelay: `${Math.min(i, 8) * 40}ms` }}
                className="group relative flex animate-rise-sm items-center gap-3.5 overflow-hidden rounded-[17px] border border-hair bg-panel p-4 shadow-glass transition-all hover:-translate-y-0.5 hover:border-hair-hi hover:bg-panel-hi"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-[14px] font-semibold text-fg">{project.name}</span>
                    <StatusPill tone="info">{project.framework}</StatusPill>
                  </div>
                  {project.repository_url && (
                    <p className="mt-1 truncate font-mono text-[11px] text-fg3">{project.repository_url}</p>
                  )}
                </div>
                <ChevronRight className="h-4 w-4 flex-none text-fg3" aria-hidden="true" />
              </Link>
            ))}
            {meta?.has_more && (
              <div className="pt-3.5 text-center">
                <button
                  onClick={loadMore}
                  disabled={loading}
                  className="h-9 rounded-[13px] border border-hair-hi bg-panel px-4 text-[12.5px] font-semibold text-fg2 transition-colors hover:bg-panel-hi disabled:opacity-50"
                >
                  {loading ? "Loading…" : "Load More"}
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

function EmptyPanel({ icon: Icon, title, description }: { icon: typeof Bot; title: string; description: string }) {
  return (
    <div className="flex animate-pop flex-col items-center justify-center gap-2 rounded-[22px] border border-dashed border-hair-hi bg-panel p-16 text-center shadow-glass">
      <div className="mb-1.5 flex h-[62px] w-[62px] animate-float-y items-center justify-center rounded-[20px] bg-acc-soft text-acc">
        <Icon className="h-6 w-6" aria-hidden="true" />
      </div>
      <h3 className="m-0 text-[19px] font-semibold tracking-tight text-fg">{title}</h3>
      <p className="m-0 max-w-sm text-[13px] text-fg2">{description}</p>
    </div>
  );
}

function ListSkeleton({ count }: { count: number }) {
  return (
    <div className="flex flex-col gap-2">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="h-[86px] animate-pulse rounded-[17px] border border-hair bg-panel" />
      ))}
    </div>
  );
}
