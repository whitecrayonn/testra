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
  status: "draft" | "active" | "deprecated";
  priority: "low" | "medium" | "high" | "critical";
  tags: string[];
  version: number;
  created_by: string;
  created_at: string;
  updated_at: string;
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
