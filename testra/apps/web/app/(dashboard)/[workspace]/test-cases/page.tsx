"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { Search, Plus, ChevronRight, TestTube } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { CardSkeleton } from "@/components/ui/skeleton";
import { LinkButton } from "@/components/ui/link-button";
import { listTestCases, searchTestCases } from "@/features/testmanagement/api";
import type { TestCase, PaginationMeta } from "@/types/testmanagement";

const statusVariants: Record<string, "neutral" | "success" | "danger"> = {
  draft: "neutral",
  active: "success",
  deprecated: "danger",
};

const priorityVariants: Record<string, "neutral" | "info" | "warning" | "danger"> = {
  low: "neutral",
  medium: "info",
  high: "warning",
  critical: "danger",
};

export default function TestCasesPage() {
  const [cases, setCases] = useState<TestCase[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchMode, setSearchMode] = useState(false);
  const [cursor, setCursor] = useState<string | undefined>(undefined);
  const [error, setError] = useState<string | null>(null);

  const projectId =
    typeof window !== "undefined"
      ? localStorage.getItem("testra_project_id") || ""
      : "";
  const workspaceId =
    typeof window !== "undefined"
      ? localStorage.getItem("testra_workspace_id") || ""
      : "";

  const fetchCases = useCallback(
    async (reset?: boolean) => {
      setLoading(true);
      setError(null);
      try {
        if (searchMode && searchQuery.trim()) {
          const result = await searchTestCases(workspaceId, searchQuery, {
            cursor: reset ? undefined : cursor,
          });
          setCases((prev) => (reset ? result.data : [...prev, ...result.data]));
          setMeta(result.meta);
        } else if (projectId) {
          const result = await listTestCases(projectId, {
            cursor: reset ? undefined : cursor,
          });
          setCases((prev) => (reset ? result.data : [...prev, ...result.data]));
          setMeta(result.meta);
        } else {
          setCases([]);
          setError("No project selected. Select a project first.");
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load test cases");
      } finally {
        setLoading(false);
      }
    },
    [cursor, projectId, searchMode, searchQuery, workspaceId],
  );

  useEffect(() => {
    fetchCases(true);
  }, [fetchCases]);

  const handleSearch = () => {
    setSearchMode(searchQuery.trim().length > 0);
    setCursor(undefined);
    setCases([]);
    setTimeout(() => fetchCases(true), 0);
  };

  const handleLoadMore = () => {
    if (meta?.next_cursor) {
      setCursor(meta.next_cursor);
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Test Cases"
        description="Manage and search your test case repository."
        actions={
          <LinkButton href="/dashboard/test-cases/new">
            <Plus className="mr-2 h-4 w-4" aria-hidden="true" />
            New Test Case
          </LinkButton>
        }
      />

      <div className="flex flex-col gap-2 sm:flex-row">
        <Input
          placeholder="Search test cases..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSearch()}
          className="max-w-md"
        />
        <Button variant="secondary" onClick={handleSearch}>
          <Search className="mr-2 h-4 w-4" aria-hidden="true" />
          Search
        </Button>
      </div>

      {error && (
        <div role="alert">
          <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">
            {error}
          </Card>
        </div>
      )}

      <div className="space-y-2">
        {!projectId ? (
          <EmptyState
            icon={TestTube}
            title="No project selected"
            description="Select a project in Projects to view and create test cases."
            action={{ label: "Go to Projects", href: "/dashboard/projects" }}
          />
        ) : loading && cases.length === 0 ? (
          <CardSkeleton count={4} />
        ) : cases.length === 0 ? (
          <EmptyState
            icon={TestTube}
            title="No test cases found"
            description="Create your first test case to get started."
            action={{ label: "New Test Case", href: "/dashboard/test-cases/new" }}
            secondaryAction={{ label: "Go to Projects", href: "/dashboard/projects" }}
          />
        ) : (
          <>
            {cases.map((tc) => (
              <Link
                key={tc.id}
                href={`/dashboard/test-cases/${tc.id}`}
                className="group block"
              >
                <Card className="flex items-center justify-between p-4 transition-colors group-hover:border-brand-300">
                  <div className="flex-1 space-y-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="font-medium text-slate-900">{tc.title}</span>
                      <Badge variant={statusVariants[tc.status] || "neutral"}>{tc.status}</Badge>
                      <Badge variant={priorityVariants[tc.priority] || "neutral"}>{tc.priority}</Badge>
                    </div>
                    {tc.description && (
                      <p className="line-clamp-1 text-sm text-slate-500">{tc.description}</p>
                    )}
                    <div className="flex flex-wrap items-center gap-3 text-xs text-slate-400">
                      <span>v{tc.version}</span>
                      <span>{tc.steps.length} steps</span>
                      {tc.tags.length > 0 && <span>{tc.tags.join(", ")}</span>}
                    </div>
                  </div>
                  <ChevronRight className="h-5 w-5 text-slate-400" aria-hidden="true" />
                </Card>
              </Link>
            ))}
            {meta?.has_more && (
              <div className="pt-4 text-center">
                <Button variant="secondary" onClick={handleLoadMore} loading={loading}>
                  Load More
                </Button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
