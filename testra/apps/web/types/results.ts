export interface TestRun {
  id: string;
  workspace_id: string;
  project_id: string;
  suite_id: string | null;
  name: string;
  status: "pending" | "running" | "passed" | "failed" | "skipped" | "cancelled";
  total: number;
  passed: number;
  failed: number;
  skipped: number;
  blocked: number;
  duration_ms: number;
  source: "manual" | "ci" | "api";
  created_by: string;
  started_at: string | null;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface TestRunItem {
  id: string;
  run_id: string;
  test_case_id: string | null;
  title: string;
  status: "pending" | "running" | "passed" | "failed" | "skipped" | "blocked";
  duration_ms: number;
  error_message: string;
  stack_trace: string;
  artifacts: string[];
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface PaginationMeta {
  next_cursor: string | null;
  has_more: boolean;
}

interface PaginatedResponse<T> {
  data: T[];
  meta: PaginationMeta;
}

export interface RunProgressEvent {
  run_id: string;
  item_id: string;
  status: string;
  total: number;
  passed: number;
  failed: number;
  skipped: number;
  blocked: number;
  progress: number;
}
