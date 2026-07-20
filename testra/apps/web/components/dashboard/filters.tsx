"use client";

import { useCallback, useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
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
    "w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100";

  return (
    <Card className="p-4 dark:bg-slate-900 dark:border-slate-700">
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-700 dark:text-slate-300">Release</label>
          <input
            className={inputClass}
            placeholder="e.g. v1.2.0"
            value={filter.release || ""}
            onChange={(e) => setFilter((f) => ({ ...f, release: e.target.value }))}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-700 dark:text-slate-300">Sprint</label>
          <input
            className={inputClass}
            placeholder="e.g. Sprint 3"
            value={filter.sprint || ""}
            onChange={(e) => setFilter((f) => ({ ...f, sprint: e.target.value }))}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-700 dark:text-slate-300">Environment</label>
          <input
            className={inputClass}
            placeholder="e.g. staging"
            value={filter.environment || ""}
            onChange={(e) => setFilter((f) => ({ ...f, environment: e.target.value }))}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-700 dark:text-slate-300">Source</label>
          <select
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
          <label className="mb-1 block text-xs font-medium text-slate-700 dark:text-slate-300">Start</label>
          <input
            type="date"
            className={inputClass}
            value={filter.start || ""}
            onChange={(e) => setFilter((f) => ({ ...f, start: e.target.value || undefined }))}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-700 dark:text-slate-300">End</label>
          <input
            type="date"
            className={inputClass}
            value={filter.end || ""}
            onChange={(e) => setFilter((f) => ({ ...f, end: e.target.value || undefined }))}
          />
        </div>
      </div>
      <div className="mt-3 flex justify-end gap-2">
        <Button variant="secondary" onClick={clear}>
          Clear
        </Button>
        <Button onClick={apply}>Apply</Button>
      </div>
    </Card>
  );
}
