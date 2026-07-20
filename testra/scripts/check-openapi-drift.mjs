import { spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

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

function readOpenApiPaths() {
  const content = readFileSync(openApiPath, "utf8");
  const paths = new Set();

  const lines = content.split(/\r?\n/);
  let currentPath = null;
  for (const line of lines) {
    const pathMatch = line.match(/^( {2,4})\/[^\s].*?:\s*$/);
    if (pathMatch) {
      currentPath = line.trim().replace(/:$/, "");
      continue;
    }
    if (!currentPath) continue;

    const methodMatch = line.match(/^( {4,6})(get|post|put|patch|delete|head|options|trace):\s*$/i);
    if (methodMatch) {
      const method = methodMatch[2].toUpperCase();
      paths.add(`${method} ${currentPath}`);
    }
  }
  return paths;
}

const generated = generateRoutes();
const manual = readOpenApiPaths();

const missing = [];
const wellKnown = new Set(["GET /.well-known/jwks.json", "GET /health"]);

for (const route of generated) {
  const relative = stripApiV1(route.path);
  const key = `${route.method} ${relative}`;
  if (wellKnown.has(key)) continue;
  if (!manual.has(key)) {
    missing.push(key);
  }
}

if (missing.length > 0) {
  console.error("OpenAPI drift detected: the following routes are implemented but not documented:");
  for (const key of missing) {
    console.error(`  ${key}`);
  }
  process.exit(1);
}

console.log(`OpenAPI is synchronized with the chi router (${generated.length} routes checked).`);
