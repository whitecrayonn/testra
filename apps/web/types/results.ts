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

export type RunItemStatus = "pending" | "running" | "passed" | "failed" | "skipped" | "blocked" | "retest" | "not_executed";

export interface StepResult {
  order: number;
  status: RunItemStatus;
  comment: string;
  duration_ms: number;
  executed_by?: string | null;
  executed_at?: string | null;
}

export interface Evidence {
  id: string;
  run_item_id: string;
  step_order: number;
  file_name: string;
  content_type: string;
  storage_path: string;
  uploaded_by?: string | null;
  created_at: string;
}

export interface ExecutionHistoryEntry {
  id: string;
  run_item_id: string;
  status: RunItemStatus;
  step_results: StepResult[];
  comment: string;
  duration_ms: number;
  executed_by?: string | null;
  created_at: string;
}

export interface TestRunItem {
  id: string;
  run_id: string;
  test_case_id: string | null;
  title: string;
  status: RunItemStatus;
  duration_ms: number;
  error_message: string;
  stack_trace: string;
  artifacts: string[];
  step_results: StepResult[];
  comment: string;
  executed_by?: string | null;
  executed_at?: string | null;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface TestPlan {
  id: string;
  workspace_id: string;
  project_id: string;
  suite_id: string | null;
  name: string;
  description: string;
  status: "active" | "archived";
  configuration: Record<string, unknown>;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface TestPlanItem {
  id: string;
  plan_id: string;
  test_case_id: string;
  sort_order: number;
  created_at: string;
}

export interface PaginationMeta {
  next_cursor: string | null;
  has_more: boolean;
}

export interface PaginatedResponse<T> {
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
