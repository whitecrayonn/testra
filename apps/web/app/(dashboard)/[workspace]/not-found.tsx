"use client";

import { ErrorState } from "@/components/error-state";

export default function WorkspaceNotFound() {
  return <ErrorState type="workspace" />;
}
