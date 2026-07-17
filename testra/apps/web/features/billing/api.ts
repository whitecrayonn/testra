import { apiFetch } from "@/lib/api";
import type { Subscription, Invoice } from "@/types/billing";

export async function getSubscription(workspaceId: string): Promise<Subscription> {
  return apiFetch(`/api/v1/billing/subscription?workspace_id=${workspaceId}`);
}

export async function updateSubscription(input: {
  workspace_id: string;
  plan: string;
  seats: number;
  cancel_at_period_end?: boolean;
}): Promise<Subscription> {
  return apiFetch("/api/v1/billing/subscription", {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function listInvoices(workspaceId: string): Promise<Invoice[]> {
  return apiFetch(`/api/v1/billing/invoices?workspace_id=${workspaceId}`);
}
