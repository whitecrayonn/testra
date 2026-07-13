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

const verifySchema = z.object({
  code: z.string().min(6, "Enter the 6-digit code").max(6),
});

type VerifyValues = z.infer<typeof verifySchema>;

interface MFASetupResponse {
  secret: string;
  qr_code: string;
}

export default function MFASetupPage() {
  const router = useRouter();
  const [setupData, setSetupData] = useState<MFASetupResponse | null>(null);
  const [serverError, setServerError] = useState("");
  const [enabled, setEnabled] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<VerifyValues>({
    resolver: zodResolver(verifySchema),
  });

  const handleSetup = async () => {
    setServerError("");
    try {
      const data = await apiFetch<MFASetupResponse>("/api/v1/auth/mfa/setup", {
        method: "POST",
      });
      setSetupData(data);
    } catch (err) {
      if (err instanceof ApiError) {
        setServerError(err.message);
      }
    }
  };

  const onVerify = async (values: VerifyValues) => {
    setServerError("");
    try {
      await apiFetch("/api/v1/auth/mfa/verify", {
        method: "POST",
        body: JSON.stringify({ code: values.code }),
      });
      setEnabled(true);
      setTimeout(() => router.push("/dashboard"), 1500);
    } catch (err) {
      if (err instanceof ApiError) {
        setServerError(err.message);
      }
    }
  };

  if (enabled) {
    return (
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>MFA Enabled</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-slate-600">
            Multi-factor authentication is now enabled. Redirecting...
          </p>
        </CardContent>
      </Card>
    );
  }

  if (!setupData) {
    return (
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Set up MFA</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-slate-600">
            Protect your account with TOTP-based multi-factor authentication. You&apos;ll need an authenticator app like Google Authenticator, Authy, or 1Password.
          </p>
          {serverError && (
            <p className="text-sm text-red-600" role="alert">
              {serverError}
            </p>
          )}
          <Button onClick={handleSetup} className="w-full">
            Generate MFA secret
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Verify MFA</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <p className="text-sm font-medium text-slate-700">Scan this QR code:</p>
          <div className="rounded-lg border border-slate-200 bg-slate-50 p-4 text-center">
            <p className="font-mono text-sm break-all text-slate-600">{setupData.qr_code}</p>
          </div>
        </div>
        <div className="space-y-2">
          <p className="text-sm font-medium text-slate-700">Or enter manually:</p>
          <div className="rounded-lg border border-slate-200 bg-slate-50 p-3">
            <p className="font-mono text-sm break-all text-slate-600">{setupData.secret}</p>
          </div>
        </div>
        <form onSubmit={handleSubmit(onVerify)} className="space-y-4">
          <Input
            label="Enter 6-digit code"
            type="text"
            maxLength={6}
            placeholder="123456"
            error={errors.code?.message}
            {...register("code")}
          />
          {serverError && (
            <p className="text-sm text-red-600" role="alert">
              {serverError}
            </p>
          )}
          <Button type="submit" className="w-full" loading={isSubmitting}>
            Verify and enable
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
