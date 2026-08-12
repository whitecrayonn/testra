export interface Subscription {
  id: string;
  workspace_id: string;
  provider_subscription_id?: string;
  plan: string;
  status: string;
  seats: number;
  current_period_start?: string;
  current_period_end?: string;
  cancel_at_period_end: boolean;
  created_at: string;
  updated_at: string;
}

export interface Invoice {
  id: string;
  workspace_id: string;
  provider_invoice_id?: string;
  amount_cents: number;
  currency: string;
  status: string;
  period_start?: string;
  period_end?: string;
  created_at: string;
}
