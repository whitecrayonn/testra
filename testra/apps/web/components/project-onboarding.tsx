"use client";

import { FolderKanban, Plus, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";

interface ProjectOnboardingProps {
  onCreate?: () => void;
}

export function ProjectOnboarding({ onCreate }: ProjectOnboardingProps) {
  return (
    <div className="rounded-2xl border border-dashed border-slate-300 bg-slate-50 p-8 text-center dark:border-slate-700 dark:bg-slate-900">
      <div className="mx-flex mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-brand-100 text-brand-600">
        <FolderKanban className="h-6 w-6" />
      </div>
      <h2 className="mt-4 text-lg font-semibold text-slate-900 dark:text-white">
        Create your first project
      </h2>
      <p className="mx-auto mt-2 max-w-md text-sm text-slate-600 dark:text-slate-400">
        Projects group your test cases, plans, runs, and defects. Create one to start building your test suite.
      </p>
      <ul className="mx-auto mt-6 inline-flex flex-col gap-2 text-left text-sm text-slate-600 dark:text-slate-400">
        <li className="flex items-center gap-2">
          <ArrowRight className="h-4 w-4 text-brand-500" />
          Organize test cases and runs by product area
        </li>
        <li className="flex items-center gap-2">
          <ArrowRight className="h-4 w-4 text-brand-500" />
          Track defects and automation results together
        </li>
        <li className="flex items-center gap-2">
          <ArrowRight className="h-4 w-4 text-brand-500" />
          Share reports and analytics with your team
        </li>
      </ul>
      <div className="mt-8">
        <Button onClick={onCreate} size="lg">
          <Plus className="mr-2 h-4 w-4" />
          Create first project
        </Button>
      </div>
    </div>
  );
}
