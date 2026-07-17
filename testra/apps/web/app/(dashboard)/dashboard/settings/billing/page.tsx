import { CreditCard } from "lucide-react";
import { PlaceholderPage } from "@/components/ui/placeholder-page";

export default function BillingSettingsPage() {
  return (
    <PlaceholderPage
      title="Billing"
      description="Manage your subscription, payment method, and invoice history."
      icon={CreditCard}
      status="Planned for Phase 5"
      releasePhase="Phase 5 — Dashboard, Analytics & Launch"
      breadcrumbs={[
        { label: "Dashboard", href: "/dashboard" },
        { label: "Settings", href: "/dashboard/settings" },
        { label: "Billing" },
      ]}
      features={[
        { label: "Current plan and usage" },
        { label: "Payment method management" },
        { label: "Invoice history" },
        { label: "Usage and seat limits" },
      ]}
      primaryCta={{ label: "Upgrade plan" }}
      secondaryCta={{ label: "Contact support", href: "/dashboard" }}
    />
  );
}
