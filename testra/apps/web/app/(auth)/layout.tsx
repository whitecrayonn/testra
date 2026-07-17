"use client";

import Link from "next/link";
import { RouteGuard } from "@/components/auth/route-guard";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <RouteGuard requireAuth={false} redirectTo="/dashboard">
      <div className="flex min-h-screen flex-col items-center justify-center bg-slate-50 px-4">
        <div className="mb-8 text-center">
          <Link href="/" className="text-2xl font-bold text-brand-600">
            Testra
          </Link>
          <p className="mt-1 text-sm text-slate-500">One Platform. Every Test.</p>
        </div>
        {children}
      </div>
    </RouteGuard>
  );
}
