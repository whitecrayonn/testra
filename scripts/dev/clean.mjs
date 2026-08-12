import { spawnSync } from "node:child_process";
import { rmSync, existsSync } from "node:fs";
import { resolve, join } from "node:path";

const repoRoot = resolve(import.meta.dirname, "..", "..");

console.log("=== Cleaning build artifacts ===\n");

const targets = [
  join(repoRoot, "apps", "web", ".next"),
  join(repoRoot, "apps", "web", ".turbo"),
  join(repoRoot, "apps", "api", "bin"),
  join(repoRoot, "apps", "api", "tmp"),
  join(repoRoot, "apps", "api", ".turbo"),
  join(repoRoot, "apps", "worker", "bin"),
  join(repoRoot, "apps", "worker", ".turbo"),
  join(repoRoot, "apps", "ml", ".turbo"),
  join(repoRoot, ".turbo"),
];

for (const target of targets) {
  if (existsSync(target)) {
    console.log(`  removing ${target.replace(repoRoot, ".")}`);
    rmSync(target, { recursive: true, force: true });
  }
}

console.log("\n=== Clean complete ===\n");
