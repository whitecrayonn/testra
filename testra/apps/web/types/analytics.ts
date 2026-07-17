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
