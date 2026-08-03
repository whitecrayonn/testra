"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useToast } from "@/lib/hooks/use-toast";
import { slugify } from "@/lib/utils";
import { createOrganization, createWorkspace, getCurrentUser, listOrganizations } from "@/features/platform/api";
import type { Organization, User } from "@/types/platform";

export function CreateWorkspaceForm() {
  const router = useRouter();
  const { toast } = useToast();
  const [user, setUser] = useState<User | null>(null);
  const [orgs, setOrgs] = useState<Organization[]>([]);
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [description, setDescription] = useState("");
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      try {
        const [me, orgList] = await Promise.all([getCurrentUser(), listOrganizations()]);
        setUser(me);
        setOrgs(orgList);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load account");
      } finally {
        setFetching(false);
      }
    }
    load();
  }, []);

  useEffect(() => {
    setSlug(slugify(name));
  }, [name]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      let organizationId = orgs[0]?.id;
      if (!organizationId) {
        const org = await createOrganization({
          name: `${name} Organization`,
          slug: slugify(`${name}-org`),
        });
        organizationId = org.id;
      }
      const workspace = await createWorkspace({
        organization_id: organizationId,
        name: name.trim(),
        slug: slug.trim() || slugify(name.trim()),
        description: description.trim(),
      });
      if (typeof window !== "undefined") {
        localStorage.setItem("testra_workspace_id", workspace.id);
        localStorage.setItem("testra_workspace_name", workspace.name);
        localStorage.setItem("testra_organization_id", workspace.organization_id);
        localStorage.removeItem("testra_project_id");
        localStorage.removeItem("testra_project_name");
      }
      toast("Workspace created", "success");
      router.replace("/dashboard");
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to create workspace";
      setError(msg);
      toast(msg, "error");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="w-full max-w-lg rounded-2xl bg-white p-8 shadow-sm ring-1 ring-slate-200 dark:bg-slate-900 dark:ring-slate-800">
      <h1 className="text-2xl font-semibold text-slate-900 dark:text-white">Create your workspace</h1>
      <p className="mt-2 text-slate-600 dark:text-slate-400">
        {user ? `Welcome, ${user.name}. ` : ""}Set up the workspace where your team will manage tests.
      </p>
      <form onSubmit={onSubmit} className="mt-8 space-y-5">
        <Input
          label="Workspace name"
          placeholder="e.g. Acme QA"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
          minLength={2}
          maxLength={100}
          disabled={loading || fetching}
        />
        <Input
          label="Slug"
          placeholder="acme-qa"
          value={slug}
          onChange={(e) => setSlug(slugify(e.target.value))}
          required
          disabled={loading || fetching}
        />
        <Input
          label="Description (optional)"
          placeholder="What is this workspace for?"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          disabled={loading || fetching}
        />
        {error && <p className="text-sm text-red-600">{error}</p>}
        <Button type="submit" size="lg" className="w-full" loading={loading || fetching} disabled={fetching}>
          Create workspace
        </Button>
      </form>
    </div>
  );
}
