export interface PaginationMeta {
  next_cursor?: string;
  has_more: boolean;
}

export interface AutomationProject {
  id: string;
  workspace_id: string;
  project_id?: string;
  name: string;
  framework: string;
  repository_url: string;
  branch: string;
  command: string;
  created_at: string;
  updated_at: string;
}

export interface AutomationExecution {
  id: string;
  project_id: string;
  workspace_id: string;
  test_run_id?: string;
  name: string;
  status: string;
  report_format: string;
  report_path?: string;
  retry_of?: string;
  duration_ms: number;
  total: number;
  passed: number;
  failed: number;
  skipped: number;
  blocked: number;
  created_at: string;
  updated_at: string;
}

export interface AutomationArtifact {
  id: string;
  execution_id: string;
  test_run_item_id?: string;
  kind: string;
  name: string;
  mime_type: string;
  file_size: number;
  metadata: Record<string, unknown>;
  created_at: string;
}

export interface AutomationLog {
  id: string;
  level: string;
  message: string;
  logged_at: string;
  created_at: string;
}

export interface IngestResult {
  run_id: string;
  execution_id?: string;
  total: number;
  passed: number;
  failed: number;
  skipped: number;
  duration_ms: number;
}
