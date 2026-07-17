import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";

export default function SettingsOverviewPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="General settings"
        description="Overview of your account and workspace configuration."
        breadcrumbs={[{ label: "Dashboard", href: "/dashboard" }, { label: "Settings" }]}
      />

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Account</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm text-slate-600">
            <p>Manage your profile, security preferences, and notification settings.</p>
            <p>Use the tabs above to navigate to each section.</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Workspace</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm text-slate-600">
            <p>Configure workspace name, members, roles, and API keys.</p>
            <p>Billing and audit logs are also available in this section.</p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
