import { ClipboardList } from "lucide-react";
import { PlaceholderPage } from "@/components/ui/placeholder-page";

export default function AuditLogsSettingsPage() {
  return (
    <PlaceholderPage
      title="Audit Logs"
      description="Review the immutable history of security and data changes in your workspace."
      icon={ClipboardList}
      status="Planned for Phase 4"
      releasePhase="Phase 4 — API Testing & Defects"
      breadcrumbs={[
        { label: "Dashboard", href: "/dashboard" },
        { label: "Settings", href: "/dashboard/settings" },
        { label: "Audit Logs" },
      ]}
      features={[
        { label: "Filter by user, resource, and action" },
        { label: "Export audit trail" },
        { label: "Immutable event log" },
        { label: "Retention policy display" },
      ]}
      primaryCta={{ label: "Export logs" }}
      secondaryCta={{ label: "Back to workspace", href: "/dashboard/settings/workspace" }}
    />
  );
}
