import { apiFetch } from "@/lib/api";
import type { AuditEvent } from "@/types/audit";

export interface PaginatedAuditEvents {
  data: AuditEvent[];
  meta: {
    has_more: boolean;
    next_cursor?: string;
  };
}

// Returns the current user's own audit trail. Scoped to the requesting user,
// not the whole organization/workspace — see the operation description on
// GET /audit-events in docs/api/openapi/openapi.yaml for why.
export async function listMyAuditEvents(cursor?: string, limit = 20): Promise<PaginatedAuditEvents> {
  const params = new URLSearchParams({ limit: String(limit) });
  if (cursor) params.set("cursor", cursor);
  return apiFetch<PaginatedAuditEvents>(`/api/v1/audit-events?${params.toString()}`);
}
