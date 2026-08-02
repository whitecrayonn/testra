"use client";

import { forwardRef, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { LogIn } from "lucide-react";

import { apiFetch, ApiError } from "@/lib/api";
import { cn } from "@/lib/utils";

const loginSchema = z.object({
  email: z.string().email("Enter a valid email"),
  password: z.string().min(1, "Password is required"),
  mfa_code: z.string().optional(),
});

type LoginValues = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const router = useRouter();
  const [serverError, setServerError] = useState("");
  const [mfaRequired, setMfaRequired] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (values: LoginValues) => {
    setServerError("");
    try {
      await apiFetch<{ token: string; refresh_token: string; user: { id: string; email: string; name: string } }>(
        "/api/v1/auth/login",
        {
          method: "POST",
          body: JSON.stringify({
            email: values.email,
            password: values.password,
            mfa_code: values.mfa_code || "",
          }),
        },
      );
      router.push("/dashboard");
    } catch (err) {
      if (err instanceof ApiError && err.code === "MFA_REQUIRED") {
        setMfaRequired(true);
        setServerError("Please enter your MFA code.");
      } else if (err instanceof ApiError) {
        setServerError(err.message);
      } else {
        setServerError("An unexpected error occurred.");
      }
    }
  };

  return (
    <div className="w-full max-w-[380px] animate-pop rounded-[22px] border border-hair bg-glass p-7 shadow-glass backdrop-blur-[26px] backdrop-saturate-150">
      <h2 className="m-0 mb-5 text-[19px] font-bold tracking-tight text-fg">Sign in to Testra</h2>
      <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4">
        <Field
          label="Email"
          type="email"
          placeholder="you@company.com"
          error={errors.email?.message}
          {...register("email")}
        />
        <Field
          label="Password"
          type="password"
          placeholder="••••••••"
          error={errors.password?.message}
          {...register("password")}
        />
        {mfaRequired && (
          <Field
            label="MFA Code"
            type="text"
            placeholder="123456"
            maxLength={6}
            error={errors.mfa_code?.message}
            {...register("mfa_code")}
          />
        )}
        {serverError && (
          <p className="text-[12.5px] text-fail" role="alert">
            {serverError}
          </p>
        )}
        <button
          type="submit"
          disabled={isSubmitting}
          aria-busy={isSubmitting || undefined}
          className="mt-1 flex h-10 items-center justify-center gap-2 rounded-[13px] bg-gradient-to-br from-acc to-acc2 text-[13px] font-bold text-white shadow-[0_10px_26px_-12px_var(--ring)] transition-transform hover:-translate-y-px disabled:pointer-events-none disabled:opacity-60"
        >
          <LogIn className="h-3.5 w-3.5" aria-hidden="true" />
          Sign in
        </button>
      </form>
      <div className="mt-5 flex flex-col gap-2 text-center text-[12.5px] text-fg2">
        <Link href="/forgot-password" className="text-acc hover:underline">
          Forgot your password?
        </Link>
        <span>
          Don&apos;t have an account?{" "}
          <Link href="/register" className="font-medium text-acc hover:underline">
            Sign up
          </Link>
        </span>
      </div>
    </div>
  );
}

interface FieldProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
}

const Field = forwardRef<HTMLInputElement, FieldProps>(({ label, error, id, className, ...props }, ref) => {
  const inputId = id || label.toLowerCase().replace(/\s+/g, "-");
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={inputId} className="text-[12.5px] font-medium text-fg2">
        {label}
      </label>
      <input
        ref={ref}
        id={inputId}
        className={cn(
          "h-10 rounded-xl border border-hair bg-panel px-3 text-[13px] text-fg placeholder-fg3 outline-none transition-colors focus:border-hair-hi focus:bg-panel-hi",
          error && "border-fail",
          className,
        )}
        {...props}
      />
      {error && <p className="text-[11.5px] text-fail">{error}</p>}
    </div>
  );
});
Field.displayName = "Field";
