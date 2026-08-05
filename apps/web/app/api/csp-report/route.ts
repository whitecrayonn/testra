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
// (not an arbitrarily large/nested value), truncate what actually gets
// written to the log, and rate-limit per client so one caller can't flood
// logs by volume of requests alone. This in-memory limiter is a
// single-process best effort (this app runs as a supervised native process,
// not serverless — see BIBLICAL_TESTRA.md "Deployment shape"), not a
// substitute for reverse-proxy/WAF-level protection in front of it.
const MAX_BODY_BYTES = 8 * 1024;
const MAX_LOGGED_CHARS = 2000;
const RATE_LIMIT_WINDOW_MS = 60_000;
const RATE_LIMIT_MAX_REQUESTS = 20;
const RATE_LIMIT_MAX_TRACKED_CLIENTS = 1000;

const rateLimitBuckets = new Map<string, { count: number; windowStart: number }>();

function clientKey(request: Request): string {
  const forwardedFor = request.headers.get("x-forwarded-for");
  if (forwardedFor) {
    return forwardedFor.split(",")[0]!.trim();
  }
  const realIP = request.headers.get("x-real-ip");
  if (realIP) {
    return realIP.trim();
  }
  // No reverse-proxy header present (e.g. local dev without one in front):
  // fall back to a single shared bucket rather than being unbounded.
  return "unknown";
}

function isRateLimited(key: string, now: number): boolean {
  const bucket = rateLimitBuckets.get(key);
  if (!bucket || now - bucket.windowStart >= RATE_LIMIT_WINDOW_MS) {
    // Bound the map itself the same way the SSRF DNS cache is bounded:
    // prune expired buckets before inserting, and skip tracking a brand
    // new client if still full afterward, rather than growing unbounded.
    for (const [k, v] of rateLimitBuckets) {
      if (now - v.windowStart >= RATE_LIMIT_WINDOW_MS) {
        rateLimitBuckets.delete(k);
      }
    }
    if (!rateLimitBuckets.has(key) && rateLimitBuckets.size >= RATE_LIMIT_MAX_TRACKED_CLIENTS) {
      return false;
    }
    rateLimitBuckets.set(key, { count: 1, windowStart: now });
    return false;
  }
  bucket.count += 1;
  return bucket.count > RATE_LIMIT_MAX_REQUESTS;
}

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
  if (isRateLimited(clientKey(request), Date.now())) {
    return new NextResponse(null, { status: 429 });
  }

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
