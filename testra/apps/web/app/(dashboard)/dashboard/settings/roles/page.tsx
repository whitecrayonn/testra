import { ShieldCheck } from "lucide-react";
import { PlaceholderPage } from "@/components/ui/placeholder-page";

export default function RolesSettingsPage() {
  return (
    <PlaceholderPage
      title="Roles"
      description="Define and assign roles and permissions for your organization."
      icon={ShieldCheck}
      status="Planned for Phase 4"
      releasePhase="Phase 4 — API Testing & Defects"
      breadcrumbs={[
        { label: "Dashboard", href: "/dashboard" },
        { label: "Settings", href: "/dashboard/settings" },
        { label: "Roles" },
      ]}
      features={[
        { label: "View built-in roles" },
        { label: "Custom role creation" },
        { label: "Permission matrix" },
        { label: "Assign roles to members" },
      ]}
      primaryCta={{ label: "Create role" }}
      secondaryCta={{ label: "Back to members", href: "/dashboard/settings/members" }}
    />
  );
}
