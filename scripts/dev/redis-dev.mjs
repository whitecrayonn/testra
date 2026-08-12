import { createConnection } from "node:net";
import { spawn, spawnSync } from "node:child_process";
import { existsSync, mkdirSync } from "node:fs";
import { resolve } from "node:path";
import { setTimeout as sleep } from "node:timers/promises";
import { tmpdir } from "node:os";

const DEFAULT_PORT = 6379;

function getPort() {
  if (process.env.REDIS_PORT) {
    return parseInt(process.env.REDIS_PORT, 10);
  }
  if (process.env.REDIS_ADDR) {
    const parts = process.env.REDIS_ADDR.split(":");
    const last = parts[parts.length - 1];
    if (last) {
      return parseInt(last, 10);
    }
  }
  return DEFAULT_PORT;
}

const REDIS_PORT = getPort();

function findExecutable(name, fallbackPaths) {
  const envVar = `${name.toUpperCase().replace(/-/g, "_")}_PATH`;
  if (process.env[envVar] && existsSync(process.env[envVar])) {
    return process.env[envVar];
  }

  const finder = process.platform === "win32" ? "where" : "which";
  const result = spawnSync(finder, [name], { encoding: "utf-8" });
  if (result.status === 0) {
    const first = result.stdout.trim().split(/\r?\n/)[0];
    if (first) {
      return first;
    }
  }

  for (const p of fallbackPaths) {
    if (existsSync(p)) {
      return p;
    }
  }
  return null;
}

function findRedisServer() {
  return findExecutable("redis-server", [
    "C:\\Program Files\\Redis\\redis-server.exe",
  ]);
}

function findRedisCli() {
  return findExecutable("redis-cli", [
    "C:\\Program Files\\Redis\\redis-cli.exe",
  ]);
}

function isPortOpen(port, timeoutMs = 1000) {
  return new Promise((resolve) => {
    let settled = false;
    const socket = createConnection(port, "127.0.0.1");
    const timer = setTimeout(() => {
      if (!settled) {
        settled = true;
        socket.destroy();
        resolve(false);
      }
    }, timeoutMs);
    socket.on("connect", () => {
      if (!settled) {
        settled = true;
        clearTimeout(timer);
        socket.destroy();
        resolve(true);
      }
    });
    socket.on("error", () => {
      if (!settled) {
        settled = true;
        clearTimeout(timer);
        resolve(false);
      }
    });
  });
}

async function waitForRedis(timeoutMs = 30000) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const open = await isPortOpen(REDIS_PORT, 1000);
    if (open) {
      return true;
    }
    await sleep(250);
  }
  return false;
}

async function waitForPortFree(timeoutMs = 5000) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const open = await isPortOpen(REDIS_PORT, 500);
    if (!open) {
      return true;
    }
    await sleep(250);
  }
  return false;
}

function pingRedis(cliPath, port) {
  const result = spawnSync(cliPath, ["-p", String(port), "ping"], {
    encoding: "utf-8",
    timeout: 1000,
  });
  return result.status === 0 && result.stdout.trim() === "PONG";
}

function shutdownRedis(cliPath, port) {
  const result = spawnSync(cliPath, ["-p", String(port), "SHUTDOWN"], {
    encoding: "utf-8",
    timeout: 5000,
  });
  return result.status === 0;
}

async function ensurePortFree(cliPath) {
  if (!(await isPortOpen(REDIS_PORT, 1000))) {
    return true;
  }

  console.log(
    `[redis-dev] Port ${REDIS_PORT} is already in use. Attempting graceful shutdown...`,
  );
  if (cliPath && pingRedis(cliPath, REDIS_PORT)) {
    shutdownRedis(cliPath, REDIS_PORT);
  }
  return waitForPortFree(5000);
}

let cleaned = false;
let redisChild = null;

function cleanup() {
  if (cleaned) {
    return;
  }
  cleaned = true;
  if (!redisChild || !redisChild.pid) {
    return;
  }
  if (process.platform === "win32") {
    spawnSync("taskkill", ["/PID", String(redisChild.pid), "/T", "/F"], {
      stdio: "ignore",
    });
  } else {
    try {
      redisChild.kill("SIGTERM");
    } catch {}
  }
}

process.on("SIGINT", () => {
  cleanup();
  process.exit(0);
});

process.on("SIGTERM", () => {
  cleanup();
  process.exit(0);
});

process.on("exit", cleanup);

async function main() {
  if (
    process.env.TESTRA_REDIS_EXTERNAL === "1" ||
    process.env.TESTRA_REDIS_EXTERNAL === "true"
  ) {
    console.log(
      "[redis-dev] TESTRA_REDIS_EXTERNAL is set; not managing Redis.",
    );
    await new Promise(() => {});
    return;
  }

  const serverPath = findRedisServer();
  const cliPath = findRedisCli();
  if (!serverPath) {
    // Redis is optional for local development: the API falls back to an
    // in-memory rate limiter. Exiting non-zero here would take down the whole
    // `pnpm dev` session over a dependency that is not required to run Testra.
    console.warn("[redis-dev] redis-server not found — continuing without Redis.");
    console.warn("[redis-dev] The API will use its in-memory rate limiter instead.");
    console.warn("[redis-dev] To enable Redis, install it or set REDIS_SERVER_PATH.");
    await new Promise(() => {});
    return;
  }

  const portFree = await ensurePortFree(cliPath);
  if (!portFree) {
    console.error(
      `[redis-dev] Could not free port ${REDIS_PORT}. Stop the existing Redis first.`,
    );
    process.exit(1);
  }

  const dataDir = resolve(tmpdir(), "testra-redis-dev");
  mkdirSync(dataDir, { recursive: true });

  const args = [
    "--port",
    String(REDIS_PORT),
    "--save",
    "",
    "--appendonly",
    "no",
    "--dir",
    dataDir,
  ];
  if (process.env.REDIS_PASSWORD) {
    args.push("--requirepass", process.env.REDIS_PASSWORD);
  }

  console.log(`[redis-dev] Starting Redis on port ${REDIS_PORT}...`);
  redisChild = spawn(serverPath, args, {
    stdio: "inherit",
    windowsHide: false,
  });

  redisChild.on("exit", (code, signal) => {
    console.log(
      `[redis-dev] Redis exited (code=${code ?? "?"}, signal=${signal ?? "?"})`,
    );
    process.exit(code ?? 0);
  });

  const ready = await waitForRedis(30000);
  if (!ready) {
    console.error(
      `[redis-dev] Redis did not become ready on port ${REDIS_PORT}.`,
    );
    cleanup();
    process.exit(1);
  }

  if (cliPath && !pingRedis(cliPath, REDIS_PORT)) {
    console.error(
      `[redis-dev] Redis is listening on ${REDIS_PORT} but is not responding to PING.`,
    );
    cleanup();
    process.exit(1);
  }

  console.log(`[redis-dev] Redis ready on port ${REDIS_PORT}.`);

  await new Promise(() => {});
}

main().catch((err) => {
  console.error("[redis-dev] Unexpected error:", err);
  cleanup();
  process.exit(1);
});
