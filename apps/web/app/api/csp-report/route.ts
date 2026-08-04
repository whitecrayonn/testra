import { NextResponse } from "next/server";

// Receives Content-Security-Policy violation reports (SBL-024). Browsers
// POST these as application/csp-report or application/reports+json; the
// body is JSON either way. This is intentionally just a log sink for MVP —
// no persistence, no external forwarding (No External LLM / customer-owns-data).
//
// This endpoint is unauthenticated by design (browsers send these reports
// without credentials), so it accepts POSTs from anyone on the internet.
// Bound the damage a hostile caller can do: cap the body size (both via a
// declared Content-Length and via a hard read limit, since Content-Length
// can be absent or wrong), require the parsed body to be a JSON object
// (not an arbitrarily large/nested value), and truncate what actually gets
// written to the log. Rate limiting is left to the reverse-proxy/WAF layer
// (see BIBLICAL_TESTRA.md "CDN / WAF future"), consistent with how other
// public, unauthenticated endpoints in this app are handled.
const MAX_BODY_BYTES = 8 * 1024;
const MAX_LOGGED_CHARS = 2000;

async function readLimitedBody(request: Request, maxBytes: number): Promise<string | null> {
  const reader = request.body?.getReader();
  if (!reader) {
    return "";
  }

  const decoder = new TextDecoder();
  let received = 0;
  let text = "";

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }
    received += value.byteLength;
    if (received > maxBytes) {
      await reader.cancel();
      return null;
    }
    text += decoder.decode(value, { stream: true });
  }

  text += decoder.decode();
  return text;
}

export async function POST(request: Request) {
  const declaredLength = Number(request.headers.get("content-length") ?? "0");
  if (Number.isFinite(declaredLength) && declaredLength > MAX_BODY_BYTES) {
    return new NextResponse(null, { status: 413 });
  }

  const raw = await readLimitedBody(request, MAX_BODY_BYTES);
  if (raw === null) {
    return new NextResponse(null, { status: 413 });
  }

  let report: unknown;
  try {
    report = JSON.parse(raw);
  } catch {
    return new NextResponse(null, { status: 400 });
  }

  if (typeof report !== "object" || report === null || Array.isArray(report)) {
    return new NextResponse(null, { status: 400 });
  }

  const serialized = JSON.stringify(report).slice(0, MAX_LOGGED_CHARS);
  console.warn("[csp-report]", serialized);

  return new NextResponse(null, { status: 204 });
}
