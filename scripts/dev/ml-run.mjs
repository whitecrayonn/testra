import { spawn } from "node:child_process";
import { existsSync } from "node:fs";
import { resolve } from "node:path";

// Runs a python module inside the ML service venv.
// Usage: node ../../scripts/dev/ml-run.mjs <module> [args...]
// Example: node ../../scripts/dev/ml-run.mjs ruff check .

const cwd = process.cwd();
const venvPython = resolve(cwd, ".venv", "bin", "python");
const venvPythonWin = resolve(cwd, ".venv", "Scripts", "python.exe");

let pythonBin = "python";

if (existsSync(venvPython)) {
  pythonBin = venvPython;
} else if (existsSync(venvPythonWin)) {
  pythonBin = venvPythonWin;
} else {
  console.warn("[ml-run] No .venv found in apps/ml.");
  console.warn('[ml-run] Run: pip install -e "apps/ml[dev]"');
  console.warn(
    "[ml-run] Or let 'pnpm install' set it up automatically via postinstall.",
  );
}

const [module, ...rest] = process.argv.slice(2);

if (!module) {
  console.error("[ml-run] No module specified.");
  process.exit(1);
}

const args = ["-m", module, ...rest];
console.log(`[ml-run] ${pythonBin} ${args.join(" ")}`);

const child = spawn(pythonBin, args, { stdio: "inherit", shell: true });
child.on("exit", (code) => process.exit(code ?? 1));
