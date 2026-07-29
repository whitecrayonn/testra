"use client";

import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { listOrganizations, listWorkspaces, getCurrentUser } from "@/features/platform/api";
import type { Organization, User, Workspace } from "@/types/platform";

interface WorkspaceContextValue {
  user: User | null;
  organization: Organization | null;
  workspace: Workspace | null;
  workspaces: Workspace[];
  loading: boolean;
  error: string | null;
  setWorkspace: (w: Workspace) => void;
  refresh: () => Promise<void>;
}

export const WorkspaceContext = createContext<WorkspaceContextValue | null>(null);

const skipRedirectPaths = ["/create-workspace", "/onboarding", "/onboarding/workspace", "/login", "/register"];

function shouldSkipRedirect(pathname: string) {
  return skipRedirectPaths.some((p) => pathname === p || pathname.startsWith(`${p}/`));
}

export function WorkspaceProvider({ children }: { children: ReactNode }) {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [workspace, setWorkspaceState] = useState<Workspace | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const setWorkspace = useCallback((w: Workspace) => {
    if (typeof window !== "undefined") {
      localStorage.setItem("testra_workspace_id", w.id);
      localStorage.setItem("testra_workspace_name", w.name);
      localStorage.setItem("testra_organization_id", w.organization_id);
    }
    const org = workspaces.find((x) => x.organization_id === w.organization_id);
    if (org) {
      setOrganization(org);
    }
    setWorkspaceState(w);
  }, [workspaces]);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const me = await getCurrentUser();
      setUser(me);
      const orgs = await listOrganizations();
      let allWorkspaces: Workspace[] = [];
      for (const org of orgs) {
        const wsList = await listWorkspaces(org.id);
        allWorkspaces = allWorkspaces.concat(wsList);
      }
      setWorkspaces(allWorkspaces);

      const storedId = typeof window !== "undefined" ? localStorage.getItem("testra_workspace_id") : null;
      const selected = allWorkspaces.find((w) => w.id === storedId) || allWorkspaces[0] || null;
      if (selected) {
        setWorkspaceState(selected);
        const org = orgs.find((o) => o.id === selected.organization_id) || orgs[0] || null;
        setOrganization(org);
        if (typeof window !== "undefined") {
          localStorage.setItem("testra_workspace_id", selected.id);
          localStorage.setItem("testra_workspace_name", selected.name);
          if (org) localStorage.setItem("testra_organization_id", org.id);
        }
      } else if (typeof window !== "undefined" && !shouldSkipRedirect(window.location.pathname)) {
        router.replace("/create-workspace");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load workspace");
    } finally {
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <WorkspaceContext.Provider
      value={{ user, organization, workspace, workspaces, loading, error, setWorkspace, refresh: load }}
    >
      {children}
    </WorkspaceContext.Provider>
  );
}

export function useWorkspace(): WorkspaceContextValue {
  const ctx = useContext(WorkspaceContext);
  if (!ctx) throw new Error("useWorkspace must be used within WorkspaceProvider");
  return ctx;
}
