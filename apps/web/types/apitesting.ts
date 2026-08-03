export interface PaginationMeta {
  next_cursor: string | null;
  has_more: boolean;
}

export interface KeyValuePair {
  key: string;
  value: string;
  enabled: boolean;
}

export interface APICollection {
  id: string;
  workspace_id: string;
  name: string;
  description: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface APIFolder {
  id: string;
  workspace_id: string;
  collection_id: string;
  parent_id: string | null;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface APIEnvironment {
  id: string;
  workspace_id: string;
  name: string;
  variables: KeyValuePair[];
  created_at: string;
  updated_at: string;
}

export type HTTPMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "HEAD" | "OPTIONS";
export type AuthType = "none" | "bearer" | "basic" | "api_key";
export type BodyType = "none" | "json" | "raw" | "form" | "urlencoded";

export interface AuthConfig {
  bearer_token?: string;
  username?: string;
  password?: string;
  api_key?: string;
  api_value?: string;
  api_location?: "header" | "query";
}

export interface APIRequest {
  id: string;
  workspace_id: string;
  collection_id: string;
  folder_id: string | null;
  environment_id: string | null;
  name: string;
  method: HTTPMethod;
  url: string;
  headers: KeyValuePair[];
  query_params: KeyValuePair[];
  auth_type: AuthType;
  auth_config: AuthConfig;
  body_type: BodyType;
  body_content: string;
  variables: KeyValuePair[];
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface APIRequestHistory {
  id: string;
  workspace_id: string;
  request_id: string | null;
  environment_id: string | null;
  name: string;
  method: string;
  url: string;
  request_headers: KeyValuePair[];
  request_body: string;
  response_status: number;
  response_status_text: string;
  response_headers: KeyValuePair[];
  response_body: string;
  response_time_ms: number;
  error: string;
  created_by: string;
  created_at: string;
}

export interface ExecutionResult {
  status: number;
  status_text: string;
  headers: Record<string, string[]>;
  body: string;
  response_time_ms: number;
  error: string;
}

export interface ExecutionResponse {
  history?: APIRequestHistory;
  result: ExecutionResult;
}
