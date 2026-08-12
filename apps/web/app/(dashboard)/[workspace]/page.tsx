"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { LayoutDashboard } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/ui/page-header";

export default function WorkspacePage() {
  const params = useParams();
  const router = useRouter();
  const workspaceId = params.workspace as string;
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    if (typeof window !== "undefined" && workspaceId) {
      localStorage.setItem("testra_workspace_id", workspaceId);
    }
  }, [workspaceId]);

  if (!mounted) {
    return <div className="p-8 text-center text-slate-500">Loading workspace...</div>;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Workspace"
        description="Workspace-specific landing page. Use the dashboard for the full overview."
      />

      <Card>
        <CardHeader>
          <CardTitle>Workspace selected</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-slate-600">
            Workspace context has been saved for this session.
          </p>
          <code className="block rounded-lg bg-slate-50 p-3 text-xs text-slate-700">{workspaceId}</code>
          <Button onClick={() => router.push("/dashboard")}>
            <LayoutDashboard className="mr-2 h-4 w-4" aria-hidden="true" />
            Go to Dashboard
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
