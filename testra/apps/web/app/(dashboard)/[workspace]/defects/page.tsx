"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { Bug, Plus, ChevronRight, FolderKanban } from "lucide-react";
import { StatusPill, type StatusTone } from "@/components/ui/status-pill";
import { listDefects, createDefect } from "@/features/defects/api";
import type { Defect, PaginationMeta } from "@/types/defects";

const STATUS_TONE: Record<string, StatusTone> = {
  open: "info",
  in_progress: "warn",
  resolved: "pass",
  closed: "neutral",
  rejected: "fail",
};

const SEVERITY_TONE: Record<string, StatusTone> = {
  low: "neutral",
  medium: "info",
  high: "warn",
  critical: "fail",
};

const SEVERITY_ACCENT: Record<string, string> = {
  low: "bg-hair-hi",
  medium: "bg-info",
  high: "bg-warn",
  critical: "bg-fail",
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
    typeof window !== "undefined" ? localStorage.getItem("testra_project_id") || "" : "";

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
    <div className="flex flex-col gap-3.5">
      <div className="flex animate-rise-sm items-end justify-between gap-4">
        <div>
          <h1 className="m-0 text-[27px] font-bold tracking-tight text-fg">Defects</h1>
          <p className="mt-1.5 text-[13px] text-fg2">Track bugs and link failures to test runs.</p>
        </div>
        <button
          onClick={() => setShowCreate((s) => !s)}
          className="flex h-[38px] items-center gap-2 rounded-[13px] bg-gradient-to-br from-acc to-acc2 px-4 text-[13px] font-bold text-white shadow-[0_10px_26px_-12px_var(--ring)] transition-transform hover:-translate-y-px"
        >
          <Plus className="h-3.5 w-3.5" aria-hidden="true" />
          {showCreate ? "Cancel" : "New Defect"}
        </button>
      </div>

      {error && (
        <div role="alert" className="rounded-[16px] border border-fail-soft bg-fail-soft p-4 text-[13px] text-fail">
          {error}
        </div>
      )}

      {showCreate && (
        <form
          onSubmit={handleCreate}
          className="flex animate-pop flex-col gap-4 rounded-[20px] border border-hair bg-panel p-5 shadow-glass"
        >
          <TextField label="Title" value={title} onChange={setTitle} placeholder="e.g. Login button unresponsive" required />
          <TextField label="Description" value={description} onChange={setDescription} placeholder="Steps to reproduce..." />
          <div className="grid gap-4 sm:grid-cols-2">
            <SelectField label="Severity" value={severity} onChange={(v) => setSeverity(v as Defect["severity"])}>
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </SelectField>
            <SelectField label="Priority" value={priority} onChange={(v) => setPriority(v as Defect["priority"])}>
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </SelectField>
          </div>
          <div className="flex gap-2.5">
            <button
              type="submit"
              disabled={creating}
              className="flex h-9 items-center gap-2 rounded-[13px] bg-gradient-to-br from-acc to-acc2 px-4 text-[12.5px] font-bold text-white disabled:pointer-events-none disabled:opacity-60"
            >
              {creating ? "Creating…" : "Create defect"}
            </button>
            <button
              type="button"
              onClick={() => setShowCreate(false)}
              className="h-9 rounded-[13px] border border-hair-hi bg-panel px-4 text-[12.5px] font-semibold text-fg transition-colors hover:bg-panel-hi"
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      <div className="flex flex-col gap-2">
        {!projectId ? (
          <EmptyPanel
            icon={FolderKanban}
            title="No project selected"
            description="Select a project from the Projects page to view and create defects."
            actionLabel="Go to Projects"
            actionHref="/dashboard/projects"
          />
        ) : loading && defects.length === 0 ? (
          <ListSkeleton count={3} />
        ) : defects.length === 0 ? (
          <EmptyPanel
            icon={Bug}
            title="No defects yet"
            description="Create your first defect to track a bug or failed test."
            actionLabel="Create first defect"
            onAction={() => setShowCreate(true)}
          />
        ) : (
          <>
            {defects.map((defect, i) => (
              <Link
                key={defect.id}
                href={`/dashboard/defects/${defect.id}`}
                style={{ animationDelay: `${Math.min(i, 8) * 40}ms` }}
                className="group relative flex animate-rise-sm items-center gap-3.5 overflow-hidden rounded-[17px] border border-hair bg-panel p-4 shadow-glass transition-all hover:-translate-y-0.5 hover:border-hair-hi hover:bg-panel-hi"
              >
                <span className={`absolute inset-y-3 left-0 w-[3px] rounded-r-[3px] ${SEVERITY_ACCENT[defect.severity] ?? "bg-hair-hi"}`} />
                <div className="min-w-0 flex-1 pl-2">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-[14px] font-semibold text-fg">{defect.title}</span>
                    <StatusPill tone={STATUS_TONE[defect.status] ?? "neutral"}>{defect.status.replace("_", " ")}</StatusPill>
                    <StatusPill tone={SEVERITY_TONE[defect.severity] ?? "neutral"}>{defect.severity}</StatusPill>
                    <StatusPill tone="neutral">{defect.priority}</StatusPill>
                  </div>
                  <div className="mt-1.5 font-mono text-[11px] text-fg3">
                    {new Date(defect.updated_at).toLocaleString()}
                  </div>
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

function TextField({
  label,
  value,
  onChange,
  placeholder,
  required,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  required?: boolean;
}) {
  const id = label.toLowerCase().replace(/\s+/g, "-");
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-[12.5px] font-medium text-fg2">
        {label}
      </label>
      <input
        id={id}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        required={required}
        className="h-10 rounded-xl border border-hair bg-panel-hi px-3 text-[13px] text-fg placeholder-fg3 outline-none transition-colors focus:border-hair-hi"
      />
    </div>
  );
}

function SelectField({
  label,
  value,
  onChange,
  children,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  children: React.ReactNode;
}) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="text-[12.5px] font-medium text-fg2">{label}</span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-10 rounded-xl border border-hair bg-panel-hi px-3 text-[13px] text-fg outline-none transition-colors focus:border-hair-hi"
      >
        {children}
      </select>
    </label>
  );
}

function EmptyPanel({
  icon: Icon,
  title,
  description,
  actionLabel,
  actionHref,
  onAction,
}: {
  icon: typeof Bug;
  title: string;
  description: string;
  actionLabel: string;
  actionHref?: string;
  onAction?: () => void;
}) {
  return (
    <div className="flex animate-pop flex-col items-center justify-center gap-2 rounded-[22px] border border-dashed border-hair-hi bg-panel p-16 text-center shadow-glass">
      <div className="mb-1.5 flex h-[62px] w-[62px] animate-float-y items-center justify-center rounded-[20px] bg-acc-soft text-acc">
        <Icon className="h-6 w-6" aria-hidden="true" />
      </div>
      <h3 className="m-0 text-[19px] font-semibold tracking-tight text-fg">{title}</h3>
      <p className="m-0 max-w-sm text-[13px] text-fg2">{description}</p>
      {actionHref ? (
        <Link
          href={actionHref}
          className="mt-3.5 flex h-[38px] items-center rounded-[13px] bg-gradient-to-br from-acc to-acc2 px-4 text-[13px] font-bold text-white"
        >
          {actionLabel}
        </Link>
      ) : (
        <button
          onClick={onAction}
          className="mt-3.5 flex h-[38px] items-center rounded-[13px] bg-gradient-to-br from-acc to-acc2 px-4 text-[13px] font-bold text-white"
        >
          {actionLabel}
        </button>
      )}
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
