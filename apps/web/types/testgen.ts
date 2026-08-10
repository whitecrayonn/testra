export interface GenerationRun {
  id: string;
  workspace_id: string;
  project_id: string;
  source: string;
  spec_filename: string;
  endpoint_count: number;
  case_count: number;
  created_by: string;
  created_at: string;
}

export interface GeneratedCaseSummary {
  id: string;
  title: string;
  priority: "low" | "medium" | "high" | "critical";
  status: string;
}

export interface GenerateFromSpecResponse {
  run: GenerationRun;
  cases: GeneratedCaseSummary[];
}

export type EndpointFieldLocation = "query" | "path" | "header" | "body";
export type EndpointFieldType = "string" | "integer" | "number" | "boolean";

export interface EndpointField {
  name: string;
  location: EndpointFieldLocation;
  type: EndpointFieldType;
  required: boolean;
  enum?: string[];
}
