"use client";

import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import type { ExecutionResult } from "@/types/apitesting";

function formatResponseBody(body: string) {
  try {
    return JSON.stringify(JSON.parse(body), null, 2);
  } catch {
    return body;
  }
}

/**
 * Renders one executed request's status/timing/body/headers — extracted
 * from api-tests/page.tsx so the test-cases "Test Now" action can show the
 * same response view without duplicating it.
 */
export function ResponseViewer({ result }: { result: ExecutionResult }) {
  const [tab, setTab] = useState<"body" | "headers">("body");

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <Badge variant={result.error || result.status >= 400 ? "danger" : "success"}>
          {result.status || "Error"} {result.status_text}
        </Badge>
        <span className="text-sm text-slate-500">{result.response_time_ms} ms</span>
        {result.error && <span className="text-sm text-red-600">{result.error}</span>}
      </div>
      <div className="flex gap-1 border-b border-slate-200">
        {(["body", "headers"] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-3 py-1.5 text-sm font-medium capitalize ${
              tab === t ? "border-b-2 border-brand-500 text-brand-700" : "text-slate-500 hover:text-slate-700"
            }`}
          >
            {t}
          </button>
        ))}
      </div>
      {tab === "body" ? (
        <pre className="max-h-96 overflow-auto rounded-lg bg-slate-900 p-4 text-sm text-slate-50">
          {formatResponseBody(result.body)}
        </pre>
      ) : (
        <pre className="max-h-96 overflow-auto rounded-lg bg-slate-50 p-4 text-sm text-slate-900">
          {JSON.stringify(result.headers, null, 2)}
        </pre>
      )}
    </div>
  );
}
