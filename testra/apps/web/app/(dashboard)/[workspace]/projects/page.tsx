"use client";

import { useState, useEffect, useCallback } from "react";
import { Folder, Plus, Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { LinkButton } from "@/components/ui/link-button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import { listProjects, createProject } from "@/features/platform/api";
import type { Project } from "@/types/platform";

function generateProjectKey(name: string): string {
  const cleaned = name.toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, 10);
  if (!cleaned) return "PROJECT";
  if (/^[0-9]/.test(cleaned)) {
    return "P" + cleaned.slice(0, 9);
  }
  if (cleaned.length < 2) return cleaned + "1";
  return cleaned;
}

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [showCreate, setShowCreate] = useState(false);
  const [name, setName] = useState("");
  const [key, setKey] = useState("");
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null);

  const workspaceId =
    typeof window !== "undefined"
      ? localStorage.getItem("testra_workspace_id") || ""
      : "";

  const fetchProjects = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      if (!workspaceId) {
        setProjects([]);
        setLoading(false);
        return;
      }
      const data = await listProjects(workspaceId);
      setProjects(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load projects");
    } finally {
      setLoading(false);
    }
  }, [workspaceId]);

  useEffect(() => {
    if (typeof window !== "undefined") {
      setSelectedProjectId(localStorage.getItem("testra_project_id"));
    }
    fetchProjects();
  }, [fetchProjects]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!workspaceId) return;
    setCreating(true);
    setError(null);
    try {
      const projectKey = key.trim() || generateProjectKey(name);
      const project = await createProject({
        workspace_id: workspaceId,
        name: name.trim(),
        key: projectKey,
      });
      setProjects((prev) => [...prev, project]);
      setName("");
      setKey("");
      setShowCreate(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create project");
    } finally {
      setCreating(false);
    }
  }

  function selectProject(project: Project) {
    if (typeof window === "undefined") return;
    localStorage.setItem("testra_project_id", project.id);
    localStorage.setItem("testra_project_name", project.name);
    setSelectedProjectId(project.id);
  }

  if (!workspaceId) {
    return (
      <div className="space-y-6">
        <PageHeader title="Projects" description="Select a workspace to view projects" />
        <EmptyState
          icon={Folder}
          title="No workspace selected"
          description="Please create or select a workspace before managing projects."
          action={{ label: "Back to Dashboard", href: "/dashboard" }}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Projects"
        description="Manage projects and select the active one for testing"
        actions={
          <Button onClick={() => setShowCreate((s) => !s)}>
            <Plus className="mr-2 h-4 w-4" />
            {showCreate ? "Cancel" : "New Project"}
          </Button>
        }
      />

      {error && (
        <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">
          {error}
        </Card>
      )}

      {showCreate && (
        <Card className="p-6">
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <Input
                label="Project name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Web Application"
                required
              />
              <Input
                label="Project key"
                value={key}
                onChange={(e) => setKey(e.target.value)}
                placeholder="WEB"
              />
            </div>
            <div className="flex gap-2">
              <Button type="submit" loading={creating}>
                Create project
              </Button>
              <Button variant="secondary" onClick={() => setShowCreate(false)}>
                Cancel
              </Button>
            </div>
          </form>
        </Card>
      )}

      {loading ? (
        <div className="space-y-3">
          <Skeleton className="h-20 w-full" count={3} />
        </div>
      ) : projects.length === 0 ? (
        <EmptyState
          icon={Folder}
          title="No projects yet"
          description="Create your first project to start organizing test cases and runs."
          action={{ label: "Create project", onClick: () => setShowCreate(true) }}
        />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {projects.map((project) => {
            const selected = project.id === selectedProjectId;
            return (
              <Card
                key={project.id}
                className={`p-5 transition-colors hover:border-brand-300 ${
                  selected ? "border-brand-300 bg-brand-50" : ""
                }`}
              >
                <div className="flex items-start justify-between">
                  <div>
                    <h3 className="font-semibold text-slate-900">{project.name}</h3>
                    <p className="text-sm text-slate-500">{project.key}</p>
                  </div>
                  {selected && (
                    <Badge variant="success">
                      <Check className="mr-1 h-3 w-3" />
                      Active
                    </Badge>
                  )}
                </div>
                {project.description && (
                  <p className="mt-2 line-clamp-2 text-sm text-slate-600">{project.description}</p>
                )}
                <div className="mt-4 flex items-center gap-2">
                  <Button
                    variant={selected ? "secondary" : "primary"}
                    size="sm"
                    onClick={() => selectProject(project)}
                  >
                    {selected ? "Selected" : "Select"}
                  </Button>
                  <LinkButton
                    href={`/dashboard/test-cases?project=${project.id}`}
                    variant="ghost"
                    size="sm"
                  >
                    View cases
                  </LinkButton>
                </div>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
