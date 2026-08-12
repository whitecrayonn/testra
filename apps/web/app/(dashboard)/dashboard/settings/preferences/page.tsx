import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PageHeader } from "@/components/ui/page-header";
import { Button } from "@/components/ui/button";

export default function PreferencesSettingsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Preferences"
        description="Customize your Testra experience."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "Preferences" },
        ]}
      />

      <Card>
        <CardHeader>
          <CardTitle>Appearance</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-slate-600">Theme and density settings will be added here.</p>
          <Button disabled>Save preferences</Button>
          <p className="text-xs text-slate-500">Preference persistence is planned for a future release.</p>
        </CardContent>
      </Card>
    </div>
  );
}
