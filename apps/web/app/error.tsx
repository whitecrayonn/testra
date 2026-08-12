"use client";

import { useEffect } from "react";
import { ErrorState } from "@/components/error-state";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // eslint-disable-next-line no-console
    console.error("Application error:", error);
  }, [error]);

  return <ErrorState type="500" reset={reset} />;
}
