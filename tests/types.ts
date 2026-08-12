export interface User {
  id: string;
  email: string;
  name: string;
  mfa_enabled?: boolean;
}

export interface AuthResult {
  token: string;
  refresh_token: string;
  user: User;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  description?: string;
}

export interface Workspace {
  id: string;
  organization_id: string;
  name: string;
  slug: string;
  description?: string;
}

export interface Project {
  id: string;
  workspace_id: string;
  name: string;
  key: string;
  description?: string;
}

export interface TestStep {
  action: string;
  expected: string;
  test_data?: string;
}

export interface TestCase {
  id: string;
  workspace_id: string;
  project_id: string;
  title: string;
  description?: string;
  preconditions?: string;
  status: "draft" | "active" | "deprecated";
  priority: "low" | "medium" | "high" | "critical";
  tags: string[];
  steps: TestStep[];
  version?: number;
}

export interface TestPlan {
  id: string;
  project_id: string;
  workspace_id: string;
  name: string;
  description?: string;
  test_case_ids: string[];
}

export interface TestRun {
  id: string;
  project_id: string;
  workspace_id: string;
  name: string;
  status: string;
  source: string;
  passed: number;
  failed: number;
  skipped: number;
  total: number;
  created_at: string;
}

export interface TestRunItem {
  id: string;
  test_run_id: string;
  test_case_id: string;
  status: string;
}

export interface Defect {
  id: string;
  project_id: string;
  workspace_id: string;
  title: string;
  description?: string;
  status: "open" | "in_progress" | "resolved" | "closed" | "rejected";
  severity: "low" | "medium" | "high" | "critical";
  priority: "low" | "medium" | "high" | "critical";
  updated_at: string;
}

export interface AutomationProject {
  id: string;
  workspace_id: string;
  project_id?: string;
  name: string;
  framework: string;
  repository_url?: string;
  branch?: string;
  command?: string;
}

export interface ApiCollection {
  id: string;
  workspace_id: string;
  name: string;
  description?: string;
}

export interface ApiEnvironment {
  id: string;
  workspace_id: string;
  name: string;
  variables: Record<string, string>;
}

export interface ApiRequest {
  id: string;
  collection_id: string;
  name: string;
  method: string;
  url: string;
  headers?: Record<string, string>;
  body?: string;
}

export interface ApiExecution {
  id: string;
  request_id: string;
  status: number;
  response_body?: string;
}

export interface Notification {
  id: string;
  user_id?: string;
  title: string;
  message: string;
  read: boolean;
  created_at: string;
}

export interface Integration {
  id: string;
  workspace_id: string;
  name: string;
  provider: string;
  enabled: boolean;
}

export interface SearchResult {
  results: Array<{
    id: string;
    type: string;
    title: string;
  }>;
}

export interface DashboardSummary {
  total_test_cases: number;
  total_test_runs: number;
  pass_rate: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  meta?: {
    next_cursor?: string;
    has_more?: boolean;
  };
}
