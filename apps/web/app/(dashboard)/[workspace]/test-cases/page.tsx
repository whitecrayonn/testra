"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { Search, Plus, ChevronRight, TestTube, FileJson, Check } from "lucide-react";
import { StatusPill } from "@/components/ui/status-pill";
import {
  listTestCases,
  searchTestCases,
  approveTestCase,
} from "@/features/testmanagement/api";
import type { TestCase, PaginationMeta } from "@/types/testmanagement";

const PRIORITY_TONE: Record<string, "fail" | "warn" | "info" | "neutral"> = {
  critical: "fail",
  high: "warn",
  medium: "info",
  low: "neutral",
};

const STATUS_TONE: Record<string, "pass" | "fail" | "warn" | "info" | "neutral"> = {
  active: "pass",
  draft: "neutral",
  deprecated: "fail",
  pending_review: "warn",
};

const PRIORITY_ACCENT: Record<string, string> = {
  critical: "bg-fail",
  high: "bg-warn",
  medium: "bg-info",
  low: "bg-hair-hi",
};

export default function TestCasesPage() {
  const [cases, setCases] = useState<TestCase[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchMode, setSearchMode] = useState(false);
  const [cursor, setCursor] = useState<string | undefined>(undefined);
  const [error, setError] = useState<string | null>(null);
  const [needsReviewOnly, setNeedsReviewOnly] = useState(false);
  const [approvingId, setApprovingId] = useState<string | null>(null);

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

  async function handleApprove(id: string) {
    setApprovingId(id);
    setError(null);
    try {
      const updated = await approveTestCase(id);
      setCases((prev) => prev.map((c) => (c.id === id ? updated : c)));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to approve test case");
    } finally {
      setApprovingId(null);
    }
  }

  const pendingReviewCount = cases.filter((c) => c.status === "pending_review").length;
  const visibleCases = needsReviewOnly
    ? cases.filter((c) => c.status === "pending_review")
    : cases;

  return (
    <div className="flex flex-col gap-3.5">
      <div className="flex animate-rise-sm items-end justify-between gap-4">
        <div>
          <h1 className="m-0 text-[27px] font-bold tracking-tight text-fg">Test Cases</h1>
          <p className="mt-1.5 text-[13px] text-fg2">Manage and search your test case repository.</p>
        </div>
        <div className="flex gap-2.5">
          <Link
            href="/dashboard/test-cases/generate"
            className="flex h-[38px] items-center gap-2 rounded-[13px] border border-hair-hi bg-panel px-4 text-[13px] font-semibold text-fg transition-colors hover:bg-panel-hi"
          >
            <FileJson className="h-3.5 w-3.5" aria-hidden="true" />
            Generate from Spec
          </Link>
          <Link
            href="/dashboard/test-cases/new"
            className="flex h-[38px] items-center gap-2 rounded-[13px] bg-gradient-to-br from-acc to-acc2 px-4 text-[13px] font-bold text-white shadow-[0_10px_26px_-12px_var(--ring)] transition-transform hover:-translate-y-px"
          >
            <Plus className="h-3.5 w-3.5" aria-hidden="true" />
            New Test Case
          </Link>
        </div>
      </div>

      <div className="flex animate-rise-sm gap-2.5">
        <div className="flex h-[38px] max-w-md flex-1 items-center gap-2 rounded-[13px] border border-hair bg-panel px-3.5">
          <Search className="h-3.5 w-3.5 text-fg3" aria-hidden="true" />
          <input
            placeholder="Search test cases..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSearch()}
            className="flex-1 bg-transparent text-[13px] text-fg outline-none placeholder:text-fg3"
          />
        </div>
        <button
          onClick={handleSearch}
          className="flex h-[38px] items-center gap-2 rounded-[13px] border border-hair-hi bg-panel px-4 text-[13px] font-semibold text-fg transition-colors hover:bg-panel-hi"
        >
          <Search className="h-3.5 w-3.5" aria-hidden="true" />
          Search
        </button>
        {pendingReviewCount > 0 && (
          <button
            onClick={() => setNeedsReviewOnly((v) => !v)}
            aria-pressed={needsReviewOnly}
            className={`flex h-[38px] items-center gap-2 rounded-[13px] border px-4 text-[13px] font-semibold transition-colors ${
              needsReviewOnly
                ? "border-warn bg-warn-soft text-warn"
                : "border-hair-hi bg-panel text-fg hover:bg-panel-hi"
            }`}
          >
            Needs Review ({pendingReviewCount})
          </button>
        )}
      </div>

      {error && (
        <div role="alert" className="rounded-[16px] border border-fail-soft bg-fail-soft p-4 text-[13px] text-fail">
          {error}
        </div>
      )}

      <div className="flex flex-col gap-2">
        {!projectId ? (
          <NoProject />
        ) : loading && cases.length === 0 ? (
          <ListSkeleton count={4} />
        ) : cases.length === 0 ? (
          <div className="flex animate-pop flex-col items-center justify-center gap-2 rounded-[22px] border border-dashed border-hair-hi bg-panel p-16 text-center shadow-glass">
            <div className="mb-1.5 flex h-[62px] w-[62px] animate-float-y items-center justify-center rounded-[20px] bg-acc-soft text-acc">
              <TestTube className="h-6 w-6" aria-hidden="true" />
            </div>
            <h3 className="m-0 text-[19px] font-semibold tracking-tight text-fg">No test cases found</h3>
            <p className="m-0 max-w-sm text-[13px] text-fg2">Create your first test case to get started.</p>
            <div className="mt-3.5 flex gap-2.5">
              <Link
                href="/dashboard/test-cases/new"
                className="flex h-[38px] items-center gap-2 rounded-[13px] bg-gradient-to-br from-acc to-acc2 px-4 text-[13px] font-bold text-white"
              >
                <Plus className="h-3.5 w-3.5" aria-hidden="true" />
                New Test Case
              </Link>
              <Link
                href="/dashboard/projects"
                className="flex h-[38px] items-center rounded-[13px] border border-hair-hi bg-panel px-4 text-[13px] font-semibold text-fg transition-colors hover:bg-panel-hi"
              >
                Go to Projects
              </Link>
            </div>
          </div>
        ) : (
          <>
            {visibleCases.map((tc, i) => (
              <Link
                key={tc.id}
                href={`/dashboard/test-cases/${tc.id}`}
                style={{ animationDelay: `${Math.min(i, 8) * 40}ms` }}
                className="group relative flex animate-rise-sm items-center gap-3.5 overflow-hidden rounded-[17px] border border-hair bg-panel p-4 shadow-glass transition-all hover:-translate-y-0.5 hover:border-hair-hi hover:bg-panel-hi"
              >
                <span className={`absolute inset-y-3 left-0 w-[3px] rounded-r-[3px] ${PRIORITY_ACCENT[tc.priority] ?? "bg-hair-hi"}`} />
                <div className="min-w-0 flex-1 pl-2">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-[14px] font-semibold text-fg">{tc.title}</span>
                    <StatusPill tone={STATUS_TONE[tc.status] ?? "neutral"}>
                      {tc.status.replace("_", " ")}
                    </StatusPill>
                    <StatusPill tone={PRIORITY_TONE[tc.priority] ?? "neutral"}>{tc.priority}</StatusPill>
                    {tc.source === "generated_spec" && (
                      <StatusPill tone="info">generated</StatusPill>
                    )}
                  </div>
                  {tc.description && (
                    <p className="mt-1 truncate text-[12.5px] text-fg3">{tc.description}</p>
                  )}
                  <div className="mt-1.5 flex flex-wrap items-center gap-3 font-mono text-[11px] text-fg3">
                    <span>v{tc.version}</span>
                    <span>{tc.steps.length} steps</span>
                    {tc.tags.length > 0 && <span>{tc.tags.join(", ")}</span>}
                  </div>
                </div>
                {tc.status === "pending_review" && (
                  <button
                    onClick={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                      handleApprove(tc.id);
                    }}
                    disabled={approvingId === tc.id}
                    className="flex h-8 flex-none items-center gap-1.5 rounded-[10px] border border-pass-soft bg-pass-soft px-3 text-[12px] font-semibold text-pass transition-colors hover:brightness-95 disabled:opacity-50"
                  >
                    <Check className="h-3.5 w-3.5" aria-hidden="true" />
                    {approvingId === tc.id ? "Approving…" : "Approve"}
                  </button>
                )}
                <ChevronRight className="h-4 w-4 flex-none text-fg3" aria-hidden="true" />
              </Link>
            ))}
            {meta?.has_more && (
              <div className="pt-3.5 text-center">
                <button
                  onClick={handleLoadMore}
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

function NoProject() {
  return (
    <div className="flex animate-pop flex-col items-center justify-center gap-2 rounded-[22px] border border-dashed border-hair-hi bg-panel p-16 text-center shadow-glass">
      <div className="mb-1.5 flex h-[62px] w-[62px] items-center justify-center rounded-[20px] bg-acc-soft text-acc">
        <TestTube className="h-6 w-6" aria-hidden="true" />
      </div>
      <h3 className="m-0 text-[19px] font-semibold tracking-tight text-fg">No project selected</h3>
      <p className="m-0 max-w-sm text-[13px] text-fg2">Select a project in Projects to view and create test cases.</p>
      <Link
        href="/dashboard/projects"
        className="mt-3.5 flex h-[38px] items-center rounded-[13px] bg-gradient-to-br from-acc to-acc2 px-4 text-[13px] font-bold text-white"
      >
        Go to Projects
      </Link>
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
