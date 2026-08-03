import { spawn, execSync } from "node:child_process";
import { existsSync } from "node:fs";
import { resolve } from "node:path";

const cmdPath = process.argv[2];
if (!cmdPath) {
  console.error("Usage: go-dev.mjs <cmd-path> (e.g. cmd/api)");
  process.exit(1);
}

const cwd = process.cwd();
const airToml = resolve(cwd, ".air.toml");

if (process.platform === "win32" && !process.env.GOCACHE) {
  const localAppData = process.env.LOCALAPPDATA || resolve(process.env.USERPROFILE || "", "AppData", "Local");
  process.env.GOCACHE = resolve(localAppData, "go-build");
}

function hasAir() {
  try {
    execSync("air -v", { stdio: "ignore", shell: true });
    return true;
  } catch {
    return false;
  }
}

if (hasAir() && existsSync(airToml)) {
  console.log("[go-dev] Using air for hot reload");
  const child = spawn("air", [], { stdio: "inherit", shell: true });
  child.on("exit", (code) => process.exit(code ?? 1));
} else {
  if (!hasAir()) {
    console.log("[go-dev] air not found — using 'go run' (no hot reload).");
    console.log("[go-dev] Install air for hot reload: go install github.com/air-verse/air@latest");
  }
  const child = spawn("go", ["run", `./${cmdPath}`], { stdio: "inherit", shell: true });
  child.on("exit", (code) => process.exit(code ?? 1));
}
