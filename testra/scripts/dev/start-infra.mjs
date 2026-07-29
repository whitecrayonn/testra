import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { resolve } from "node:path";

function run(cmd, args, opts = {}) {
  console.log(`> ${cmd} ${args.join(" ")}`);
  const result = spawnSync(cmd, args, { stdio: "inherit", shell: true, ...opts });
  if (result.status !== 0) {
    console.error(`Command failed: ${cmd} ${args.join(" ")}`);
    process.exit(result.status ?? 1);
  }
}

function hasCommand(cmd) {
  const r = spawnSync(`${cmd} --version`, { stdio: "ignore", shell: true });
  return r.status === 0;
}

function checkService(name, checkFn) {
  const ok = checkFn();
  if (!ok) {
    console.error(`  [FAIL] ${name} is not reachable.`);
    return false;
  }
  console.log(`  [OK]   ${name}`);
  return true;
}

const repoRoot = resolve(import.meta.dirname, "..", "..");

// --- 1. Check local services ---
console.log("\n=== Checking local services (native development — ADR-009) ===\n");

const goBin = "C:\\Program Files\\Go\\bin";
if (process.platform === "win32" && existsSync(goBin) && !process.env.PATH?.includes(goBin)) {
  process.env.PATH = `${goBin};${process.env.PATH}`;
}

let allOk = true;

// PostgreSQL — try connecting via psql
allOk = checkService("PostgreSQL", () => {
  if (hasCommand("psql")) {
    const r = spawnSync("psql", ["-h", "localhost", "-p", "5432", "-U", "testra", "-d", "testra", "-c", "SELECT 1"], {
      stdio: "pipe",
      shell: true,
      encoding: "utf-8",
      env: { ...process.env, PGPASSWORD: "testra" },
    });
    return r.status === 0;
  }
  console.warn("  psql not found — skipping PostgreSQL check (ensure it is running on localhost:5432)");
  return true;
}) && allOk;

// Redis is optional in MVP (the code falls back to an in-memory rate limiter),
// so we only verify PostgreSQL before continuing.

if (!allOk) {
  console.error("\nPostgreSQL is not reachable.");
  console.error("Install and start PostgreSQL 16+ and create the testra database/user.");
  console.error("See README.md for platform-specific installation instructions.");
  process.exit(1);
}

console.log("\nAll required services are reachable.");

// --- 2. Run database migrations ---
console.log("\n=== Running database migrations ===\n");

if (!hasCommand("go")) {
  console.warn("Go is not installed — skipping migrations.");
  console.warn("Install Go 1.23+ and run: go run ./apps/api/cmd/migrator");
} else {
  run("go", ["run", "./apps/api/cmd/migrator"], { cwd: repoRoot });
}

console.log("\n=== Infrastructure ready ===\n");
