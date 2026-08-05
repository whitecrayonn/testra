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
// written to the log, and rate-limit in two layers — per client, and a
// global backstop that a caller can't evade by varying its own headers
// (see isRateLimited below for why the per-client key alone isn't trustworthy).
// This in-memory limiter is a single-process best effort (this app runs as
// a supervised native process, not serverless — see BIBLICAL_TESTRA.md
// "Deployment shape"), not a substitute for reverse-proxy/WAF-level
// protection in front of it.
const MAX_BODY_BYTES = 8 * 1024;
const MAX_LOGGED_CHARS = 2000;
const RATE_LIMIT_WINDOW_MS = 60_000;
const RATE_LIMIT_MAX_REQUESTS_PER_CLIENT = 20;
const RATE_LIMIT_MAX_TRACKED_CLIENTS = 1000;
// X-Forwarded-For/X-Real-IP are client-supplied unless a trusted reverse
// proxy overwrites them (nginx/Caddy config isn't in this repo yet — see
// BIBLICAL_TESTRA.md "Deployment shape"), so a caller can trivially get a
// "fresh" per-client bucket by varying the header on every request. The
// per-client limiter below still throttles naive flooding from a single,
// header-unaware client; this global cap is a backstop that can't be
// defeated by header spoofing, since it isn't keyed by anything
// client-supplied.
const RATE_LIMIT_GLOBAL_MAX_REQUESTS = 500;

const rateLimitBuckets = new Map<string, { count: number; windowStart: number }>();
let globalBucket = { count: 0, windowStart: 0 };

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

function isGloballyRateLimited(now: number): boolean {
  if (now - globalBucket.windowStart >= RATE_LIMIT_WINDOW_MS) {
    globalBucket = { count: 1, windowStart: now };
    return false;
  }
  globalBucket.count += 1;
  return globalBucket.count > RATE_LIMIT_GLOBAL_MAX_REQUESTS;
}

function isPerClientRateLimited(key: string, now: number): boolean {
  const bucket = rateLimitBuckets.get(key);
  if (!bucket || now - bucket.windowStart >= RATE_LIMIT_WINDOW_MS) {
    // Bound the map itself the same way the SSRF DNS lookup is bounded by a
    // timeout: prune expired buckets before inserting.
    for (const [k, v] of rateLimitBuckets) {
      if (now - v.windowStart >= RATE_LIMIT_WINDOW_MS) {
        rateLimitBuckets.delete(k);
      }
    }
    if (!rateLimitBuckets.has(key) && rateLimitBuckets.size >= RATE_LIMIT_MAX_TRACKED_CLIENTS) {
      // Fail closed: with no capacity left to track a new client this
      // window, treat it as rate-limited rather than letting it through
      // unlimited. (A prior version of this returned false/"not limited"
      // here, which meant a flood of distinct or spoofed keys could bypass
      // the limiter entirely once the map filled up — the global cap above
      // is the real backstop for that case, but this should still fail
      // safe rather than open.)
      return true;
    }
    rateLimitBuckets.set(key, { count: 1, windowStart: now });
    return false;
  }
  bucket.count += 1;
  return bucket.count > RATE_LIMIT_MAX_REQUESTS_PER_CLIENT;
}

// isGloballyRateLimited always runs first (and always updates its own
// state), so the global backstop trips regardless of how many distinct
// per-client keys a caller cycles through.
function isRateLimited(key: string, now: number): boolean {
  const globallyLimited = isGloballyRateLimited(now);
  const perClientLimited = isPerClientRateLimited(key, now);
  return globallyLimited || perClientLimited;
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
