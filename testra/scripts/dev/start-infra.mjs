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

// Redis — try redis-cli ping
allOk = checkService("Redis", () => {
  if (hasCommand("redis-cli")) {
    const r = spawnSync("redis-cli", ["-h", "localhost", "-p", "6379", "ping"], {
      stdio: "pipe",
      shell: true,
      encoding: "utf-8",
    });
    return (r.stdout || "").trim() === "PONG";
  }
  console.warn("  redis-cli not found — skipping Redis check (ensure it is running on localhost:6379)");
  return true;
}) && allOk;

if (!allOk) {
  console.error("\nSome local services are not reachable.");
  console.error("Install and start: PostgreSQL 16+, Redis 7+, Mailpit, MinIO");
  console.error("See README.md for platform-specific installation instructions.");
  console.error("\nDocker Compose is available as an optional alternative:");
  console.error("  docker compose -f infra/docker/docker-compose.yml up -d");
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
