import { spawn } from "node:child_process";
import { existsSync } from "node:fs";
import { resolve } from "node:path";

const cwd = process.cwd();
const venvPython = resolve(cwd, ".venv", "bin", "python");
const venvPythonWin = resolve(cwd, ".venv", "Scripts", "python.exe");

let pythonBin = "python";

if (existsSync(venvPython)) {
  pythonBin = venvPython;
} else if (existsSync(venvPythonWin)) {
  pythonBin = venvPythonWin;
} else {
  console.warn("[ml-dev] No .venv found in apps/ml.");
  console.warn("[ml-dev] Run: pip install -e \"apps/ml[dev]\"");
  console.warn("[ml-dev] Or let 'pnpm install' set it up automatically via postinstall.");
}

const args = ["-m", "uvicorn", "api.main:app", "--reload", "--port", "8000", "--host", "0.0.0.0"];
console.log(`[ml-dev] ${pythonBin} ${args.join(" ")}`);

const child = spawn(pythonBin, args, { stdio: "inherit", shell: true });
child.on("exit", (code) => process.exit(code ?? 1));
