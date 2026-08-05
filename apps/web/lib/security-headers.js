/**
 * Pure, framework-free construction of the security response headers applied
 * by middleware.ts. Kept dependency-free (no next/server import) so it can
 * be exercised directly by node:test without pulling in the Next.js runtime.
 *
 * @param {{ isProduction: boolean, isHttps: boolean, apiBaseUrl: string }} input
 * @returns {Record<string, string>}
 */
function buildSecurityHeaders(input) {
  const headers = {};

  if (input.isProduction && input.isHttps) {
    headers["Strict-Transport-Security"] =
      "max-age=63072000; includeSubDomains; preload";
  }

  headers["X-Content-Type-Options"] = "nosniff";
  headers["X-Frame-Options"] = "SAMEORIGIN";
  headers["Referrer-Policy"] = "strict-origin-when-cross-origin";
  headers["X-DNS-Prefetch-Control"] = "off";
  headers["Permissions-Policy"] =
    "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()";

  let apiOrigin = input.apiBaseUrl;
  try {
    apiOrigin = new URL(input.apiBaseUrl).origin;
  } catch {
    apiOrigin = input.apiBaseUrl;
  }

  const isDev = !input.isProduction;
  const scriptEval = isDev ? " 'unsafe-eval'" : "";

  headers["Content-Security-Policy"] = [
    "default-src 'self'",
    `script-src 'self' 'unsafe-inline'${scriptEval}`,
    "style-src 'self' 'unsafe-inline'",
    "img-src 'self' data:",
    "font-src 'self'",
    `connect-src 'self' ${apiOrigin}`,
    "frame-ancestors 'none'",
    "object-src 'none'",
    "base-uri 'self'",
    "form-action 'self'",
    isDev ? "" : "upgrade-insecure-requests",
    "report-uri /api/csp-report",
  ]
    .filter(Boolean)
    .join("; ");

  return headers;
}

module.exports = { buildSecurityHeaders };
