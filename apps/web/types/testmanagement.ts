export interface TestStep {
  order: number;
  action: string;
  expected: string;
  test_data: string;
}

export interface TestCase {
  id: string;
  workspace_id: string;
  project_id: string;
  suite_id: string | null;
  title: string;
  description: string;
  preconditions: string;
  steps: TestStep[];
  status: "draft" | "active" | "deprecated" | "pending_review";
  priority: "low" | "medium" | "high" | "critical";
  tags: string[];
  version: number;
  created_by: string;
  created_at: string;
  updated_at: string;
  /** "manual" for user-authored cases; "generated_spec" for the deterministic
   * OpenAPI-driven generator (no AI/ML); "generated_file" for the LLM-backed
   * Excel/CSV upload generator — the one opt-in exception to "no AI/ML" (see
   * docs/BIBLICAL_TESTRA.md and apps/api/internal/testgen). */
  source: "manual" | "generated_spec" | "generated_file";
  generation_run_id: string | null;
  reviewed_by: string | null;
}

export interface TestCaseVersion {
  id: string;
  test_case_id: string;
  version: number;
  title: string;
  description: string;
  preconditions: string;
  steps: TestStep[];
  changed_by: string;
  created_at: string;
}

export interface TestFolder {
  id: string;
  workspace_id: string;
  parent_id: string | null;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface TestSuite {
  id: string;
  workspace_id: string;
  folder_id: string | null;
  name: string;
  description: string;
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
