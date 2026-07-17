"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { apiFetch, setAuth, ApiError } from "@/lib/api";

const registerSchema = z.object({
  name: z.string().min(1, "Name is required"),
  email: z.string().email("Enter a valid email"),
  password: z.string().min(12, "Password must be at least 12 characters"),
});

type RegisterValues = z.infer<typeof registerSchema>;

export default function RegisterPage() {
  const router = useRouter();
  const [serverError, setServerError] = useState("");

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterValues>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (values: RegisterValues) => {
    setServerError("");
    try {
      const data = await apiFetch<{ token: string; refresh_token: string; user: { id: string; email: string; name: string } }>(
        "/api/v1/auth/register",
        {
          method: "POST",
          body: JSON.stringify(values),
        },
      );
      setAuth(data.token, data.refresh_token);
      router.push("/onboarding");
    } catch (err) {
      if (err instanceof ApiError) {
        setServerError(err.message);
      } else {
        setServerError("An unexpected error occurred.");
      }
    }
  };

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Create your account</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <Input
            label="Full name"
            type="text"
            placeholder="Jane Doe"
            error={errors.name?.message}
            {...register("name")}
          />
          <Input
            label="Email"
            type="email"
            placeholder="you@company.com"
            error={errors.email?.message}
            {...register("email")}
          />
          <Input
            label="Password"
            type="password"
            placeholder="••••••••"
            error={errors.password?.message}
            {...register("password")}
          />
          {serverError && (
            <p className="text-sm text-red-600" role="alert">
              {serverError}
            </p>
          )}
          <Button type="submit" className="w-full" loading={isSubmitting}>
            Create account
          </Button>
        </form>
      </CardContent>
      <CardFooter className="text-sm text-slate-600">
        Already have an account?{" "}
        <Link href="/login" className="ml-1 text-brand-600 hover:underline">
          Sign in
        </Link>
      </CardFooter>
    </Card>
  );
}
