import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { resolve } from "node:path";

function run(cmd, args, opts = {}) {
  console.log(`> ${cmd} ${args.join(" ")}`);
  // Pass argv directly (no shell) so arguments are not re-parsed or concatenated.
  const result = spawnSync(cmd, args, { stdio: "inherit", shell: false, ...opts });
  if (result.status !== 0) {
    console.error(`Command failed: ${cmd} ${args.join(" ")}`);
    process.exit(result.status ?? 1);
  }
}

// Resolve the binary instead of probing a version flag. Not every tool accepts
// `--version` — `go` in particular only understands `go version`, which made the
// old check report Go as missing and silently skip database migrations.
function hasCommand(cmd) {
  const probe = process.platform === "win32" ? "where" : "command";
  const args = process.platform === "win32" ? [cmd] : ["-v", cmd];
  return spawnSync(probe, args, { stdio: "ignore", shell: false }).status === 0;
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
      shell: false,
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

// Migrations are not optional: starting the API against an unmigrated database
// fails in confusing ways at request time, so stop here rather than warn.
if (!hasCommand("go")) {
  console.error("Go is required to run database migrations, but the `go` command was not found.");
  console.error("Install Go 1.24+ from https://go.dev/dl/ and make sure it is on your PATH.");
  console.error("If Go is installed to a custom location, add its `bin` directory to PATH.");
  process.exit(1);
}

run("go", ["run", "./apps/api/cmd/migrator"], { cwd: repoRoot });

console.log("\n=== Infrastructure ready ===\n");
