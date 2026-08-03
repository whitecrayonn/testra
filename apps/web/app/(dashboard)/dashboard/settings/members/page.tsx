import { Users } from "lucide-react";
import { PlaceholderPage } from "@/components/ui/placeholder-page";

export default function MembersSettingsPage() {
  return (
    <PlaceholderPage
      title="Members"
      description="Invite and manage workspace members, assign roles, and control access."
      icon={Users}
      status="Planned for Phase 4"
      releasePhase="Phase 4 — API Testing & Defects"
      breadcrumbs={[
        { label: "Dashboard", href: "/dashboard" },
        { label: "Settings", href: "/dashboard/settings" },
        { label: "Members" },
      ]}
      features={[
        { label: "Invite members by email" },
        { label: "Role assignment" },
        { label: "Pending invitations" },
        { label: "Remove members" },
      ]}
      primaryCta={{ label: "Invite member" }}
      secondaryCta={{ label: "Back to workspace", href: "/dashboard/settings/workspace" }}
    />
  );
}
