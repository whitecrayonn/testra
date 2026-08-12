export interface Dashboard {
  id: string;
  workspace_id: string;
  name: string;
  type: string;
  config: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface Summary {
  total_runs: number;
  passed: number;
  failed: number;
  skipped: number;
  blocked: number;
  duration_ms: number;
}

export interface TrendPoint {
  date: string;
  total_runs: number;
  passed: number;
  failed: number;
  skipped: number;
  blocked: number;
  duration_ms: number;
}

export interface Metrics {
  total_test_cases: number;
  total_test_plans: number;
  total_test_runs: number;
  execution_progress: number;
  pass_rate: number;
  fail_rate: number;
  blocked: number;
  retest: number;
  skipped: number;
  automation_coverage: number;
  api_test_coverage: number;
  execution_duration_ms: number;
  average_execution_time_ms: number;
  top_failed_test_cases: TopFailedItem[];
  top_failed_suites: TopFailedSuite[];
  top_failed_apis: TopFailedAPI[];
  most_active_qa: ActiveUser[];
  most_active_automation: ActiveUser[];
  defect_density: number;
  open_defects: number;
  closed_defects: number;
  defect_aging: DefectAging;
  bug_reopen_rate: number;
  recent_activity: Activity[];
  execution_timeline: TimelinePoint[];
  weekly_trend: TrendPoint[];
  monthly_trend: TrendPoint[];
  release_quality_trend: ReleaseQualityPoint[];
}

export interface TopFailedItem {
  test_case_id: string;
  title: string;
  failures: number;
}

export interface TopFailedSuite {
  suite_id: string;
  name: string;
  failures: number;
}

export interface TopFailedAPI {
  request_id: string;
  name: string;
  failures: number;
}

export interface ActiveUser {
  user_id: string;
  name: string;
  count: number;
}

export interface DefectAging {
  average_days: number;
  max_days: number;
}

export interface Activity {
  id: string;
  type: string;
  title: string;
  created_by: string;
  created_at: string;
}

export interface TimelinePoint {
  date: string;
  total_runs: number;
  passed: number;
  failed: number;
  skipped: number;
  blocked: number;
  duration_ms: number;
}

export interface ReleaseQualityPoint {
  release: string;
  passed: number;
  failed: number;
  skipped: number;
  blocked: number;
  total: number;
}

export interface MetricsFilter {
  workspace_id: string;
  project_id?: string;
  release?: string;
  sprint?: string;
  environment?: string;
  tester?: string;
  source?: string;
  start?: string;
  end?: string;
}
