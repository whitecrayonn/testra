export const TEST_CONFIG = {
  timeouts: {
    fast: 2000,
    default: 5000,
    slow: 10000,
    navigation: 15000,
  },
  retries: {
    api: 1,
    ui: 1,
  },
  pagination: {
    defaultLimit: 20,
    maxLimit: 100,
  },
  browsers: ["chromium", "firefox", "webkit"],
  devices: ["Pixel 5", "iPhone 12"],
} as const;

export const API_BASE = process.env.TEST_API_URL || "http://localhost:8080";
export const WEB_BASE = process.env.TEST_WEB_URL || "http://localhost:3000";
export const API_V1 = `${API_BASE}/api/v1`;
