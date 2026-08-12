import { apiFetch } from "@/lib/api";

export interface SearchItem {
  id: string;
  type: string;
  title: string;
  subtitle: string;
  url: string;
}

export interface SearchResult {
  workspace_id: string;
  query: string;
  projects: SearchItem[];
  test_cases: SearchItem[];
  defects: SearchItem[];
  automation: SearchItem[];
  api_collections: SearchItem[];
  test_plans: SearchItem[];
  users: SearchItem[];
}

export async function globalSearch(workspaceId: string, query: string, limit = 10): Promise<SearchResult> {
  return apiFetch<SearchResult>(
    `/api/v1/search?workspace_id=${workspaceId}&q=${encodeURIComponent(query)}&limit=${limit}`,
  );
}
