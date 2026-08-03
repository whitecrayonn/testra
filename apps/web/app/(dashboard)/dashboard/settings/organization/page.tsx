"use client";

import { useState, useEffect } from "react";
import { Building2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/ui/page-header";
import { listOrganizations } from "@/features/platform/api";
import type { Organization } from "@/types/platform";

export default function OrganizationSettingsPage() {
  const [org, setOrg] = useState<Organization | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    listOrganizations()
      .then((orgs) => {
        if (!cancelled) setOrg(orgs[0] ?? null);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Organization"
        description="Manage your organization name and high-level settings."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "Organization" },
        ]}
      />

      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-brand-50 text-brand-600">
              <Building2 className="h-6 w-6" aria-hidden="true" />
            </div>
            <CardTitle>Organization details</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {loading ? (
            <p className="text-sm text-slate-500">Loading organization...</p>
          ) : org ? (
            <>
              <Input label="Organization name" value={org.name} disabled />
              <Input label="Slug" value={org.slug} disabled />
              <Button disabled>Save changes</Button>
              <p className="text-xs text-slate-500">Organization editing will be enabled in a future release.</p>
            </>
          ) : (
            <p className="text-sm text-slate-500">No organization found. Create one during onboarding.</p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
