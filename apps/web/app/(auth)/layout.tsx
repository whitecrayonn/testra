"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { RouteGuard } from "@/components/auth/route-guard";
import { AmbientBackground } from "@/components/dashboard/ambient-background";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const themed = pathname === "/login";

  if (themed) {
    return (
      <RouteGuard requireAuth={false} redirectTo="/dashboard">
        <AmbientBackground />
        <main className="relative z-[1] flex min-h-screen flex-col items-center justify-center px-4">
          <div className="mb-7 flex animate-rise-sm flex-col items-center gap-2.5 text-center">
            <Link
              href="/"
              aria-label="Testra home"
              className="flex h-11 w-11 items-end justify-center gap-[3px] rounded-[13px] bg-gradient-to-br from-acc to-acc2 pb-2.5 shadow-[0_10px_28px_-10px_var(--ring)]"
            >
              <span aria-hidden="true" className="h-2.5 w-1 rounded-sm bg-white/95" />
              <span aria-hidden="true" className="h-4 w-1 rounded-sm bg-white/95" />
              <span aria-hidden="true" className="h-[13px] w-1 rounded-sm bg-white/95" />
            </Link>
            <h1 className="text-[19px] font-bold tracking-[0.14em] text-fg">TESTRA</h1>
            <p className="font-mono text-[10px] tracking-[0.12em] text-fg2">ONE PLATFORM · EVERY TEST</p>
          </div>
          {children}
        </main>
      </RouteGuard>
    );
  }

  return (
    <RouteGuard requireAuth={false} redirectTo="/dashboard">
      <main className="flex min-h-screen flex-col items-center justify-center bg-slate-50 px-4">
        <div className="mb-8 text-center">
          <h1 className="text-2xl font-bold text-brand-600">
            <Link href="/">Testra</Link>
          </h1>
          <p className="mt-1 text-sm text-slate-500">One Platform. Every Test.</p>
        </div>
        {children}
      </main>
    </RouteGuard>
  );
}
