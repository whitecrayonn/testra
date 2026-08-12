/* eslint-disable @typescript-eslint/no-require-imports -- plain CommonJS so
   node:test can run it directly without a build/transpile step. */
const test = require("node:test");
const assert = require("node:assert/strict");
const { buildSecurityHeaders } = require("./security-headers");

test("sets baseline hardening headers regardless of environment", () => {
  const headers = buildSecurityHeaders({
    isProduction: false,
    isHttps: false,
    apiBaseUrl: "http://localhost:8080",
  });

  assert.equal(headers["X-Content-Type-Options"], "nosniff");
  assert.equal(headers["X-Frame-Options"], "SAMEORIGIN");
  assert.equal(headers["Referrer-Policy"], "strict-origin-when-cross-origin");
  assert.equal(headers["X-DNS-Prefetch-Control"], "off");
  assert.match(headers["Permissions-Policy"], /camera=\(\)/);
});

test("omits Strict-Transport-Security outside of https production", () => {
  const devHeaders = buildSecurityHeaders({
    isProduction: false,
    isHttps: false,
    apiBaseUrl: "http://localhost:8080",
  });
  assert.equal(devHeaders["Strict-Transport-Security"], undefined);

  const prodOverHttp = buildSecurityHeaders({
    isProduction: true,
    isHttps: false,
    apiBaseUrl: "https://api.testra.example",
  });
  assert.equal(prodOverHttp["Strict-Transport-Security"], undefined);
});

test("sets Strict-Transport-Security only for https production traffic", () => {
  const headers = buildSecurityHeaders({
    isProduction: true,
    isHttps: true,
    apiBaseUrl: "https://api.testra.example",
  });
  assert.equal(
    headers["Strict-Transport-Security"],
    "max-age=63072000; includeSubDomains; preload",
  );
});

test("CSP connect-src reflects the configured API origin", () => {
  const headers = buildSecurityHeaders({
    isProduction: true,
    isHttps: true,
    apiBaseUrl: "https://api.testra.example/some/path",
  });
  assert.match(headers["Content-Security-Policy"], /connect-src 'self' https:\/\/api\.testra\.example/);
});

test("CSP allows 'unsafe-eval' only outside production", () => {
  const dev = buildSecurityHeaders({
    isProduction: false,
    isHttps: false,
    apiBaseUrl: "http://localhost:8080",
  });
  assert.match(dev["Content-Security-Policy"], /script-src 'self' 'unsafe-inline' 'unsafe-eval'/);

  const prod = buildSecurityHeaders({
    isProduction: true,
    isHttps: true,
    apiBaseUrl: "https://api.testra.example",
  });
  assert.doesNotMatch(prod["Content-Security-Policy"], /unsafe-eval/);
  assert.match(prod["Content-Security-Policy"], /upgrade-insecure-requests/);
});

test("CSP always reports violations to the internal report endpoint", () => {
  const headers = buildSecurityHeaders({
    isProduction: true,
    isHttps: true,
    apiBaseUrl: "https://api.testra.example",
  });
  assert.match(headers["Content-Security-Policy"], /report-uri \/api\/csp-report/);
});
