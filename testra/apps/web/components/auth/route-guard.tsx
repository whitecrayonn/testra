"use client";

import { useEffect, useState, type ReactNode } from "react";
import { useRouter, usePathname } from "next/navigation";
import { isAuthenticated } from "@/lib/api";

interface RouteGuardProps {
  children: ReactNode;
  requireAuth?: boolean;
  redirectTo?: string;
  fallback?: ReactNode;
}

export function RouteGuard({
  children,
  requireAuth = true,
  redirectTo = "/login",
  fallback = null,
}: RouteGuardProps) {
  const router = useRouter();
  const pathname = usePathname();
  const [ready, setReady] = useState(false);
  const [authenticated, setAuthenticated] = useState(false);

  useEffect(() => {
    const auth = isAuthenticated();
    setAuthenticated(auth);

    if (requireAuth && !auth) {
      const returnUrl = encodeURIComponent(pathname);
      router.replace(`${redirectTo}?returnUrl=${returnUrl}`);
    } else if (!requireAuth && auth) {
      const params = new URLSearchParams(window.location.search);
      const returnUrl = params.get("returnUrl");
      router.replace(returnUrl || redirectTo);
    } else {
      setReady(true);
    }
  }, [requireAuth, redirectTo, pathname, router]);

  if (!ready) {
    return fallback;
  }

  if (requireAuth && !authenticated) {
    return null;
  }

  if (!requireAuth && authenticated) {
    return null;
  }

  return <>{children}</>;
}
