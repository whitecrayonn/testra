import { Shield } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { Button } from "@/components/ui/button";

export default function SecuritySettingsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Security"
        description="Manage password, multi-factor authentication, and sessions."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "Security" },
        ]}
      />

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Password</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-slate-600">Change your account password.</p>
            <Button disabled>Change password</Button>
            <p className="text-xs text-slate-500">Password changes will be enabled in a future release.</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Shield className="h-5 w-5 text-brand-600" aria-hidden="true" />
              <CardTitle>Multi-factor authentication</CardTitle>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-slate-600">Add an extra layer of security with TOTP.</p>
            <Button variant="secondary" disabled>
              Configure MFA
            </Button>
            <p className="text-xs text-slate-500">MFA configuration is managed during onboarding today.</p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
