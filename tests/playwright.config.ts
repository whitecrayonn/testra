import { defineConfig, devices } from "@playwright/test";
import path from "path";
import dotenv from "dotenv";
import dns from "node:dns";

dns.setDefaultResultOrder("ipv4first");
process.env.NODE_OPTIONS = `${process.env.NODE_OPTIONS ?? ""} --dns-result-order=ipv4first`.trim();

dotenv.config({ path: path.join(__dirname, ".env") });

export default defineConfig({
  testDir: "./e2e",
  outputDir: "./reports/test-results",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  workers: process.env.CI ? 1 : undefined,
  timeout: 60000,
  expect: {
    timeout: 15000,
  },
  reporter: [
    ["list"],
    ["html", { outputFolder: "./reports/html", open: "never" }],
    ["json", { outputFile: "./reports/json/test-results.json" }],
    ["junit", { outputFile: "./reports/junit/test-results.xml" }],
  ],
  webServer: [
    {
      cwd: path.resolve(__dirname, "..", "apps", "api"),
      command: "set MIGRATIONS_PATH=migrations && bin\\api.exe",
      url: "http://localhost:8080/health",
      timeout: 120000,
      reuseExistingServer: !process.env.CI,
    },
    {
      cwd: path.resolve(__dirname, "..", "apps", "web"),
      command: "node node_modules/next/dist/bin/next start",
      url: "http://localhost:3000",
      timeout: 120000,
      reuseExistingServer: !process.env.CI,
    },
  ],
  use: {
    baseURL: process.env.TEST_WEB_URL || "http://localhost:3000",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
    actionTimeout: 15000,
    navigationTimeout: 30000,
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
    {
      name: "firefox",
      use: { ...devices["Desktop Firefox"] },
      retries: 1,
    },
    {
      name: "webkit",
      use: { ...devices["Desktop Safari"] },
      retries: 1,
    },
    {
      name: "Mobile Chrome",
      use: { ...devices["Pixel 5"] },
    },
    {
      name: "Mobile Safari",
      use: { ...devices["iPhone 12"] },
      retries: 1,
    },
  ],
});
