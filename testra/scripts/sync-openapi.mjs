import { readFileSync, writeFileSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { resolve } from "node:path";
import * as YAML from "js-yaml";

const root = process.cwd();
const openApiPath = resolve(root, "docs/api/openapi/openapi.yaml");

function generateRoutes() {
  const result = spawnSync("go", ["run", "./apps/api/cmd/openapi"], {
    cwd: root,
    encoding: "utf8",
    maxBuffer: 10 * 1024 * 1024,
  });

  if (result.status !== 0) {
    console.error(result.stderr);
    process.exit(result.status ?? 1);
  }

  const routes = [];
  const lines = result.stdout.split(/\r?\n/);
  let current = null;
  for (const line of lines) {
    const methodMatch = line.match(/^  - method: (\S+)$/);
    if (methodMatch) {
      current = { method: methodMatch[1] };
    } else if (current) {
      const pathMatch = line.match(/^    path: (.+)$/);
      if (pathMatch) {
        current.path = pathMatch[1];
        routes.push(current);
        current = null;
      }
    }
  }
  return routes;
}

function stripApiV1(path) {
  if (path.startsWith("/api/v1")) return path.slice("/api/v1".length) || "/";
  return path;
}

function tagFor(path) {
  const segments = path.split("/").filter(Boolean);
  if (segments.length === 0) return "System";
  const first = segments[0];
  return first
    .split("-")
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join("-");
}

function operationId(method, path) {
  const segments = path.split("/").filter(Boolean);
  const words = segments.map((s) => s.replace(/[{}]/g, "").replace(/-/g, "_"));
  return `${method.toLowerCase()}${words.map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join("")}`;
}

function securityFor(method, path) {
  if (path === "/ingest") {
    return [{ apiKeyAuth: [] }];
  }
  if (
    path.startsWith("/auth/") ||
    path === "/health" ||
    path === "/.well-known/jwks.json" ||
    path === "/"
  ) {
    return null;
  }
  return [{ bearerAuth: [] }];
}

function skeletonOperation(method, path) {
  const op = {
    tags: [tagFor(path)],
    operationId: operationId(method, path),
    summary: operationId(method, path),
    responses: {
      "200": { description: "OK" },
      "401": { $ref: "#/components/responses/Unauthorized" },
      "403": { $ref: "#/components/responses/Forbidden" },
    },
  };
  const sec = securityFor(method, path);
  if (sec) {
    op.security = sec;
  }
  return op;
}

function mergeResponses(existing, patch) {
  if (!existing && !patch) return undefined;
  return { ...existing, ...patch };
}

function applyPatch(target, patch) {
  if (!target || !patch) return target;
  const result = { ...target };
  for (const key of Object.keys(patch)) {
    if (key === "responses") {
      result[key] = mergeResponses(result[key], patch[key]);
    } else if (key === "parameters") {
      result[key] = patch[key];
    } else {
      result[key] = patch[key];
    }
  }
  return result;
}

function pathPatches() {
  const csrfParam = { $ref: "#/components/parameters/CSRFTokenHeader" };
  return {
    "POST /auth/mfa/setup": {
      security: [{ bearerAuth: [] }, { cookieAuth: [] }],
      parameters: [csrfParam],
      responses: { "403": { $ref: "#/components/responses/Forbidden" } },
    },
    "POST /auth/mfa/verify": {
      security: [{ bearerAuth: [] }, { cookieAuth: [] }],
      parameters: [csrfParam],
      responses: { "403": { $ref: "#/components/responses/Forbidden" } },
    },
    "POST /auth/mfa/disable": {
      security: [{ bearerAuth: [] }, { cookieAuth: [] }],
      parameters: [csrfParam],
      responses: { "403": { $ref: "#/components/responses/Forbidden" } },
    },
    "GET /test-runs/{id}/stream": {
      description:
        "Streams `RunProgressEvent` messages as server-sent events. The connection is authenticated via the `testra_access_token` httpOnly cookie using an `EventSource` with `withCredentials: true`.",
      security: [{ bearerAuth: [] }, { cookieAuth: [] }],
      parameters: [{ name: "id", in: "path", required: true, schema: { type: "string", format: "uuid" } }],
    },
  };
}

const routes = generateRoutes();
const doc = YAML.load(readFileSync(openApiPath, "utf8"));

if (!doc.paths) doc.paths = {};

const patches = pathPatches();

for (const route of routes) {
  const rel = stripApiV1(route.path);
  if (rel === "/.well-known/jwks.json" || rel === "/health") continue;

  if (!doc.paths[rel]) {
    doc.paths[rel] = {};
  }

  const method = route.method.toLowerCase();
  if (!doc.paths[rel][method]) {
    doc.paths[rel][method] = skeletonOperation(route.method, rel);
  }

  const key = `${route.method.toUpperCase()} ${rel}`;
  if (patches[key]) {
    doc.paths[rel][method] = applyPatch(doc.paths[rel][method], patches[key]);
  }
}

// Ensure CSRF parameter and cookieAuth security scheme are defined.
if (!doc.components) doc.components = {};
if (!doc.components.securitySchemes) doc.components.securitySchemes = {};
if (!doc.components.securitySchemes.cookieAuth) {
  doc.components.securitySchemes.cookieAuth = {
    type: "apiKey",
    in: "cookie",
    name: "testra_access_token",
    description: "Short-lived JWT access token stored in an httpOnly, Secure, SameSite=Lax cookie.",
  };
}
if (!doc.components.securitySchemes.bearerAuth.description) {
  doc.components.securitySchemes.bearerAuth.description =
    "JWT access token. May be supplied via the `Authorization: Bearer <token>` header or the `testra_access_token` httpOnly cookie.";
}
if (!doc.components.parameters) doc.components.parameters = {};
if (!doc.components.parameters.CSRFTokenHeader) {
  doc.components.parameters.CSRFTokenHeader = {
    name: "X-CSRF-Token",
    in: "header",
    required: true,
    description:
      "Double-submit CSRF token value that must match the `testra_csrf_token` cookie. Required for all state-changing requests that use cookie authentication, except login, registration, refresh, and password reset.",
    schema: { type: "string" },
  };
}

const out = YAML.dump(doc, {
  lineWidth: -1,
  noRefs: true,
  quotingType: '"',
  forceQuotes: false,
});

writeFileSync(openApiPath, out);
console.log(`Synchronized ${routes.length} routes into ${openApiPath}`);
