"use client";

import { useCallback, useEffect, useState } from "react";
import type { MetricsFilter } from "@/types/analytics";

interface DashboardFiltersProps {
  value: MetricsFilter;
  onChange: (filter: MetricsFilter) => void;
}

export function DashboardFilters({ value, onChange }: DashboardFiltersProps) {
  const [filter, setFilter] = useState<MetricsFilter>(value);

  useEffect(() => {
    setFilter(value);
  }, [value]);

  const apply = useCallback(() => {
    onChange(filter);
  }, [filter, onChange]);

  const clear = useCallback(() => {
    const cleared: MetricsFilter = {
      workspace_id: filter.workspace_id,
      project_id: filter.project_id,
    };
    setFilter(cleared);
    onChange(cleared);
  }, [filter.workspace_id, filter.project_id, onChange]);

  const inputClass =
    "w-full rounded-[11px] border border-hair bg-panel px-3 py-2 text-[13px] text-fg outline-none transition-colors focus:border-acc";

  return (
    <div className="animate-pop rounded-[20px] border border-hair bg-panel p-4 shadow-glass">
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <div>
          <label htmlFor="filter-release" className="mb-1 block text-[11px] font-semibold tracking-wide text-fg3">Release</label>
          <input
            id="filter-release"
            className={inputClass}
            placeholder="e.g. v1.2.0"
            value={filter.release || ""}
            onChange={(e) => setFilter((f) => ({ ...f, release: e.target.value }))}
          />
        </div>
        <div>
          <label htmlFor="filter-sprint" className="mb-1 block text-[11px] font-semibold tracking-wide text-fg3">Sprint</label>
          <input
            id="filter-sprint"
            className={inputClass}
            placeholder="e.g. Sprint 3"
            value={filter.sprint || ""}
            onChange={(e) => setFilter((f) => ({ ...f, sprint: e.target.value }))}
          />
        </div>
        <div>
          <label htmlFor="filter-environment" className="mb-1 block text-[11px] font-semibold tracking-wide text-fg3">Environment</label>
          <input
            id="filter-environment"
            className={inputClass}
            placeholder="e.g. staging"
            value={filter.environment || ""}
            onChange={(e) => setFilter((f) => ({ ...f, environment: e.target.value }))}
          />
        </div>
        <div>
          <label htmlFor="filter-source" className="mb-1 block text-[11px] font-semibold tracking-wide text-fg3">Source</label>
          <select
            id="filter-source"
            className={inputClass}
            value={filter.source || ""}
            onChange={(e) => setFilter((f) => ({ ...f, source: e.target.value || undefined }))}
          >
            <option value="">All</option>
            <option value="manual">Manual</option>
            <option value="automation">Automation</option>
            <option value="api">API</option>
          </select>
        </div>
        <div>
          <label htmlFor="filter-start" className="mb-1 block text-[11px] font-semibold tracking-wide text-fg3">Start</label>
          <input
            id="filter-start"
            type="date"
            className={inputClass}
            value={filter.start || ""}
            onChange={(e) => setFilter((f) => ({ ...f, start: e.target.value || undefined }))}
          />
        </div>
        <div>
          <label htmlFor="filter-end" className="mb-1 block text-[11px] font-semibold tracking-wide text-fg3">End</label>
          <input
            id="filter-end"
            type="date"
            className={inputClass}
            value={filter.end || ""}
            onChange={(e) => setFilter((f) => ({ ...f, end: e.target.value || undefined }))}
          />
        </div>
      </div>
      <div className="mt-3 flex justify-end gap-2.5">
        <button
          onClick={clear}
          className="h-8 rounded-[11px] border border-hair-hi bg-transparent px-3.5 text-[12.5px] font-semibold text-fg2 transition-colors hover:bg-panel-hi"
        >
          Clear
        </button>
        <button
          onClick={apply}
          className="h-8 rounded-[11px] bg-gradient-to-br from-acc to-acc2 px-4 text-[12.5px] font-bold text-white shadow-[0_10px_26px_-12px_var(--ring)]"
        >
          Apply
        </button>
      </div>
    </div>
  );
}
