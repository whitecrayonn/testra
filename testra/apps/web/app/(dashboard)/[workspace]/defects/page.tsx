import { Bug } from "lucide-react";
import { PlaceholderPage } from "@/components/ui/placeholder-page";

export default function DefectsPage() {
  return (
    <PlaceholderPage
      title="Defects"
      description="Track bugs, link failures to runs and cases, and sync with issue trackers."
      icon={Bug}
      status="Planned for Phase 4"
      releasePhase="Phase 4 — API Testing & Defects"
      breadcrumbs={[{ label: "Dashboard", href: "/dashboard" }, { label: "Defects" }]}
      features={[
        { label: "Defect CRUD and lifecycle" },
        { label: "Link defects to test runs and test cases" },
        { label: "Severity and priority classification" },
        { label: "Jira / integration sync" },
        { label: "Defect board and filters" },
      ]}
      primaryCta={{ label: "Create defect", href: undefined }}
      secondaryCta={{ label: "View test runs", href: "/dashboard/test-runs" }}
    />
  );
}
