"use client";

import { useRouter } from "next/navigation";
import {
  AlertTriangle,
  Lock,
  LogIn,
  ServerOff,
  WifiOff,
  Frown,
  FolderX,
  Building2,
} from "lucide-react";
import { Button } from "@/components/ui/button";

type ErrorType = "404" | "403" | "401" | "500" | "offline" | "workspace" | "project" | "generic";

interface ErrorStateProps {
  type?: ErrorType;
  title?: string;
  description?: string;
  reset?: () => void;
}

const config: Record<ErrorType, { icon: React.ReactNode; title: string; description: string }> = {
  "404": {
    icon: <Frown className="h-12 w-12 text-slate-400" />,
    title: "Page not found",
    description: "We couldn't find the page you were looking for.",
  },
  "403": {
    icon: <Lock className="h-12 w-12 text-amber-500" />,
    title: "Access denied",
    description: "You don't have permission to view this resource.",
  },
  "401": {
    icon: <LogIn className="h-12 w-12 text-blue-500" />,
    title: "Session expired",
    description: "Please sign in again to continue.",
  },
  "500": {
    icon: <ServerOff className="h-12 w-12 text-red-500" />,
    title: "Something went wrong",
    description: "An unexpected error occurred. Please try again.",
  },
  offline: {
    icon: <WifiOff className="h-12 w-12 text-yellow-500" />,
    title: "You're offline",
    description: "Check your connection and try again.",
  },
  workspace: {
    icon: <Building2 className="h-12 w-12 text-slate-500" />,
    title: "Workspace not found",
    description: "The workspace you requested doesn't exist or you don't have access.",
  },
  project: {
    icon: <FolderX className="h-12 w-12 text-slate-500" />,
    title: "Project not found",
    description: "The project you requested doesn't exist or has been removed.",
  },
  generic: {
    icon: <AlertTriangle className="h-12 w-12 text-red-500" />,
    title: "Something went wrong",
    description: "An unexpected error occurred. Please try again.",
  },
};

export function ErrorState({ type = "generic", title, description, reset }: ErrorStateProps) {
  const router = useRouter();
  const c = config[type];

  return (
    <div className="flex min-h-[50vh] flex-col items-center justify-center gap-4 p-6 text-center" role="alert" aria-live="assertive">
      {c.icon}
      <h2 className="text-2xl font-semibold text-slate-900 dark:text-white">{title || c.title}</h2>
      <p className="max-w-md text-slate-600 dark:text-slate-400">{description || c.description}</p>
      <div className="flex flex-wrap items-center justify-center gap-3">
        <Button onClick={() => router.push("/dashboard")}>Go to Dashboard</Button>
        {type === "401" && (
          <Button variant="secondary" onClick={() => router.push("/login")}>
            Sign in
          </Button>
        )}
        {reset && (
          <Button variant="secondary" onClick={reset}>
            Try again
          </Button>
        )}
      </div>
    </div>
  );
}
