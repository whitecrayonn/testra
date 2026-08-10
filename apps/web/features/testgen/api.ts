import { apiFetch } from "@/lib/api";
import type { EndpointField, GenerateFromFileResponse, GenerateFromSpecResponse } from "@/types/testgen";

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

/**
 * Generates draft test cases from an uploaded .csv/.xlsx file. Unlike
 * generateFromSpec/generateFromEndpoint, this path calls an external LLM
 * (via the ML service) to interpret loosely-structured rows — the one
 * intentional exception to this feature's "no AI/ML" principle, documented
 * in docs/BIBLICAL_TESTRA.md. Rows the model can't confidently map come
 * back in skipped_rows rather than being fabricated. Fails with a clear
 * error if the backend has no ML service configured.
 */
export async function generateFromFile(input: {
  workspace_id: string;
  project_id: string;
  file: File;
  context?: string;
}): Promise<GenerateFromFileResponse> {
  const formData = new FormData();
  formData.append("workspace_id", input.workspace_id);
  formData.append("project_id", input.project_id);
  formData.append("file", input.file);
  if (input.context) {
    formData.append("context", input.context);
  }
  return apiFetch("/api/v1/generate/from-file", {
    method: "POST",
    body: formData,
  });
}
