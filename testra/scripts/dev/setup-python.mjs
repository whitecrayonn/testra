import { execSync, spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { resolve, join } from "node:path";

const repoRoot = resolve(import.meta.dirname, "..", "..");
const mlDir = join(repoRoot, "apps", "ml");
const venvDir = join(mlDir, ".venv");

function hasCommand(cmd) {
  try {
    execSync(`${cmd} --version`, { stdio: "ignore", shell: true });
    return true;
  } catch {
    return false;
  }
}

let pythonBin = null;

if (hasCommand("python")) {
  pythonBin = "python";
} else if (hasCommand("python3")) {
  pythonBin = "python3";
}

if (!pythonBin) {
  console.warn("[setup-python] Python not found — skipping ML venv setup.");
  console.warn("[setup-python] Install Python 3.12+ to enable the ML service.");
  process.exit(0);
}

// Check Python version >= 3.12
let versionOk = false;
try {
  const out = execSync(`${pythonBin} --version`, { encoding: "utf-8", shell: true }).trim();
  const match = out.match(/Python (\d+)\.(\d+)/);
  if (match) {
    const major = parseInt(match[1]);
    const minor = parseInt(match[2]);
    versionOk = major > 3 || (major === 3 && minor >= 12);
  }
} catch {}

if (!versionOk) {
  console.warn("[setup-python] Python 3.12+ required — skipping ML venv setup.");
  try {
    const ver = execSync(`${pythonBin} --version`, { encoding: "utf-8", shell: true }).trim();
    console.warn(`[setup-python] Found: ${ver}`);
  } catch {}
  process.exit(0);
}

console.log("[setup-python] Creating virtual environment for ML service...");
execSync(`${pythonBin} -m venv .venv`, { cwd: mlDir, stdio: "inherit", shell: true });

// Determine venv python path
const venvPython = process.platform === "win32"
  ? join(venvDir, "Scripts", "python.exe")
  : join(venvDir, "bin", "python");

if (!existsSync(venvPython)) {
  console.warn("[setup-python] venv creation failed — skipping.");
  process.exit(0);
}

console.log("[setup-python] Installing ML dependencies...");
spawnSync(venvPython, ["-m", "pip", "install", "-e", ".[dev]"], {
  cwd: mlDir,
  stdio: "inherit",
  shell: true,
});

console.log("[setup-python] ML service ready.");
