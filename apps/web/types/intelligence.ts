export interface RunHistoryPoint {
  status: string;
  duration_ms: number;
  date: string;
}

export interface FlakyPrediction {
  id: string;
  workspace_id: string;
  test_case_id: string;
  test_case_title: string;
  flakiness_score: number;
  confidence: number;
  explanation: string;
  recommended_action: string;
  created_at: string;
}

export interface FailureCluster {
  id: string;
  workspace_id: string;
  label: string;
  signature: string;
  count: number;
  created_at: string;
}

export interface FailureClassification {
  label: string;
  confidence: number;
  explanation: string;
}
