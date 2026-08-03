export interface AuditEvent {
  id: string;
  action: string;
  resource: string;
  resource_id?: string;
  ip_address?: string;
  metadata?: Record<string, string>;
  created_at: string;
}
