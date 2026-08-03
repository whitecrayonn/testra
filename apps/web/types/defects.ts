export type DefectStatus = "open" | "in_progress" | "resolved" | "closed" | "rejected";

export type DefectSeverity = "low" | "medium" | "high" | "critical";

export type DefectPriority = "low" | "medium" | "high" | "critical";

export interface Defect {
  id: string;
  workspace_id: string;
  project_id: string;
  test_run_item_id: string | null;
  title: string;
  description: string;
  severity: DefectSeverity;
  priority: DefectPriority;
  status: DefectStatus;
  assigned_to: string | null;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface PaginationMeta {
  next_cursor: string | null;
  has_more: boolean;
}
