"use client";

import { useState, useEffect } from "react";
import { Briefcase } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/ui/page-header";
import { listOrganizations, listWorkspaces } from "@/features/platform/api";
import type { Workspace } from "@/types/platform";

export default function WorkspaceSettingsPage() {
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const orgs = await listOrganizations();
        const org = orgs[0];
        if (org) {
          const workspaces = await listWorkspaces(org.id);
          if (!cancelled) setWorkspace(workspaces[0] ?? null);
        }
      } catch {
        // ignore
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    load();
    return () => {
      cancelled = true;
    };
  }, []);

  const storedWorkspaceId =
    typeof window !== "undefined" ? localStorage.getItem("testra_workspace_id") : null;
  const displayWorkspace = workspace ?? (storedWorkspaceId ? ({ id: storedWorkspaceId, name: "Current workspace" } as Workspace) : null);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Workspace"
        description="Manage your workspace name and settings."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "Workspace" },
        ]}
      />

      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-brand-50 text-brand-600">
              <Briefcase className="h-6 w-6" aria-hidden="true" />
            </div>
            <CardTitle>Workspace details</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {loading ? (
            <p className="text-sm text-slate-500">Loading workspace...</p>
          ) : displayWorkspace ? (
            <>
              <Input label="Workspace name" value={displayWorkspace.name} disabled />
              <Input label="Workspace ID" value={displayWorkspace.id} disabled />
              <Button disabled>Save changes</Button>
              <p className="text-xs text-slate-500">Workspace editing will be enabled in a future release.</p>
            </>
          ) : (
            <p className="text-sm text-slate-500">No workspace found. Create one during onboarding.</p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
