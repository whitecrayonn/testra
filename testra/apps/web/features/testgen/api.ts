import { apiFetch } from "@/lib/api";
import type { GenerateFromSpecResponse } from "@/types/testgen";

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
