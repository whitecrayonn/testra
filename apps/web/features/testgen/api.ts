import { apiFetch } from "@/lib/api";
import type { EndpointField, GenerateFromSpecResponse } from "@/types/testgen";

/**
 * Generates draft test cases from an OpenAPI 3.0/3.1 spec. This is
 * deterministic (rule-based) generation — no AI/ML is called anywhere in
 * this path, per docs/BIBLICAL_TESTRA.md's "No External LLM" principle.
 * Created cases have status "pending_review" and must be approved via
 * approveTestCase (features/testmanagement/api.ts) before they count toward
 * coverage.
 */
export async function generateFromSpec(input: {
  workspace_id: string;
  project_id: string;
  spec_filename?: string;
  spec: Record<string, unknown>;
}): Promise<GenerateFromSpecResponse> {
  return apiFetch("/api/v1/generate/from-spec", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

/**
 * Generates draft test cases for a single endpoint described directly by
 * method/path/fields — no OpenAPI document required. Uses the same
 * deterministic rule set as generateFromSpec (see
 * apps/api/internal/testgen/endpoint_spec.go), just fed by a form instead of
 * an uploaded spec.
 */
export async function generateFromEndpoint(input: {
  workspace_id: string;
  project_id: string;
  method: string;
  path: string;
  fields: EndpointField[];
  requires_auth: boolean;
}): Promise<GenerateFromSpecResponse> {
  return apiFetch("/api/v1/generate/from-endpoint", {
    method: "POST",
    body: JSON.stringify(input),
  });
}
