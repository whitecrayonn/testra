import { NextResponse } from "next/server";

// Receives Content-Security-Policy violation reports (SBL-024). Browsers
// POST these as application/csp-report or application/reports+json; the
// body is JSON either way, so request.json() works regardless of the
// declared content type. This is intentionally just a log sink for MVP —
// no persistence, no external forwarding (No External LLM / customer-owns-data).
export async function POST(request: Request) {
  let report: unknown;
  try {
    report = await request.json();
  } catch {
    return new NextResponse(null, { status: 400 });
  }

  console.warn("[csp-report]", JSON.stringify(report));

  return new NextResponse(null, { status: 204 });
}
