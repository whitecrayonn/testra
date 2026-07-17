"use client";

import { useState, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { apiFetch, ApiError } from "@/lib/api";

const resetSchema = z.object({
  token: z.string().min(1, "Token is required"),
  new_password: z.string().min(12, "Password must be at least 12 characters"),
});

type ResetValues = z.infer<typeof resetSchema>;

function ResetPasswordForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [serverError, setServerError] = useState("");
  const [success, setSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ResetValues>({
    resolver: zodResolver(resetSchema),
    defaultValues: {
      token: searchParams.get("token") || "",
    },
  });

  const onSubmit = async (values: ResetValues) => {
    setServerError("");
    try {
      await apiFetch("/api/v1/auth/password-reset/confirm", {
        method: "POST",
        body: JSON.stringify(values),
      });
      setSuccess(true);
      setTimeout(() => router.push("/login"), 2000);
    } catch (err) {
      if (err instanceof ApiError) {
        setServerError(err.message);
      } else {
        setServerError("An unexpected error occurred.");
      }
    }
  };

  if (success) {
    return (
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Password reset</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-slate-600">
            Your password has been reset. Redirecting to sign in...
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Set a new password</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input label="Reset token" type="text" error={errors.token?.message} {...register("token")} />
          <Input
            label="New password"
            type="password"
            placeholder="••••••••"
            error={errors.new_password?.message}
            {...register("new_password")}
          />
          {serverError && (
            <p className="text-sm text-red-600" role="alert">
              {serverError}
            </p>
          )}
          <Button type="submit" className="w-full" loading={isSubmitting}>
            Reset password
          </Button>
        </form>
      </CardContent>
      <CardFooter className="text-sm text-slate-600">
        <Link href="/login" className="text-brand-600 hover:underline">
          Back to sign in
        </Link>
      </CardFooter>
    </Card>
  );
}

export default function ResetPasswordPage() {
  return (
    <Suspense fallback={<Card className="w-full max-w-md p-8 text-center text-slate-500">Loading...</Card>}>
      <ResetPasswordForm />
    </Suspense>
  );
}
