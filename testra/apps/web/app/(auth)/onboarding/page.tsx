"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { apiFetch, ApiError } from "@/lib/api";

const onboardingSchema = z.object({
  org_name: z.string().min(1, "Organization name is required"),
  workspace_name: z.string().min(1, "Workspace name is required"),
});

type OnboardingValues = z.infer<typeof onboardingSchema>;

export default function OnboardingPage() {
  const router = useRouter();
  const [serverError, setServerError] = useState("");
  const [step, setStep] = useState(1);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<OnboardingValues>({
    resolver: zodResolver(onboardingSchema),
  });

  const onSubmit = async (values: OnboardingValues) => {
    setServerError("");
    try {
      const org = await apiFetch<{ id: string; name: string }>("/api/v1/organizations", {
        method: "POST",
        body: JSON.stringify({ name: values.org_name }),
      });

      const workspace = await apiFetch<{ id: string; name: string }>("/api/v1/workspaces", {
        method: "POST",
        body: JSON.stringify({ organization_id: org.id, name: values.workspace_name }),
      });

      router.push(`/dashboard`);
    } catch (err) {
      if (err instanceof ApiError) {
        setServerError(err.message);
      } else {
        setServerError("An unexpected error occurred.");
      }
    }
  };

  return (
    <Card className="w-full max-w-lg">
      <CardHeader>
        <CardTitle>Welcome to Testra</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="mb-6 flex items-center gap-2 text-sm text-slate-500">
          <span className={step >= 1 ? "font-bold text-brand-600" : ""}>1. Organization</span>
          <span>→</span>
          <span className={step >= 2 ? "font-bold text-brand-600" : ""}>2. Workspace</span>
          <span>→</span>
          <span className={step >= 3 ? "font-bold text-brand-600" : ""}>3. Done</span>
        </div>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            label="Organization name"
            type="text"
            placeholder="Acme Inc."
            error={errors.org_name?.message}
            {...register("org_name")}
          />
          <Input
            label="Workspace name"
            type="text"
            placeholder="Engineering Team"
            error={errors.workspace_name?.message}
            {...register("workspace_name")}
          />
          {serverError && (
            <p className="text-sm text-red-600" role="alert">
              {serverError}
            </p>
          )}
          <Button type="submit" className="w-full" loading={isSubmitting}>
            Create and continue
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
