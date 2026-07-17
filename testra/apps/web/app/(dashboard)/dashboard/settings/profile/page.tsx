"use client";

import { useState, useEffect } from "react";
import { User as UserIcon } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/ui/page-header";
import { getCurrentUser } from "@/features/platform/api";
import type { User } from "@/types/platform";

export default function ProfileSettingsPage() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    getCurrentUser()
      .then((data) => {
        if (!cancelled) setUser(data);
      })
      .catch(() => {
        if (!cancelled) setError("Unable to load profile");
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
        title="Profile"
        description="View and update your personal information."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "Profile" },
        ]}
      />

      {error && (
        <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>
      )}

      {loading ? (
        <Card className="p-8 text-center text-slate-500">Loading profile...</Card>
      ) : user ? (
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-brand-50 text-brand-600">
                <UserIcon className="h-6 w-6" aria-hidden="true" />
              </div>
              <CardTitle>Personal details</CardTitle>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <Input label="Name" value={user.name} disabled />
            <Input label="Email" value={user.email} disabled />
            <p className="text-sm text-slate-500">Profile editing is coming soon.</p>
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}
