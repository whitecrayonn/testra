const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

let refreshing: Promise<void> | null = null;
let csrfToken: string | null = null;
let csrfPromise: Promise<string | null> | null = null;

function isBrowser(): boolean {
  return typeof window !== "undefined";
}

function redirectToLogin() {
  if (isBrowser()) {
    window.location.href = "/login";
  }
}

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
  }
}

interface ApiEnvelope<T> {
  data?: T;
  error?: { code: string; message: string };
  meta?: Record<string, unknown>;
}

function isMutatingRequest(method: string | undefined): boolean {
  switch ((method || "GET").toUpperCase()) {
    case "GET":
    case "HEAD":
    case "OPTIONS":
    case "TRACE":
      return false;
    default:
      return true;
  }
}

function shouldSendJsonBody(options: RequestInit): boolean {
  if (options.body) {
    return typeof options.body === "string";
  }
  const method = (options.method || "GET").toUpperCase();
  return method === "POST" || method === "PUT" || method === "PATCH";
}

async function ensureCsrfToken(): Promise<string | null> {
  if (csrfToken) return csrfToken;
  if (csrfPromise) return csrfPromise;

  csrfPromise = (async () => {
    try {
      const res = await fetch(`${API_URL}/api/v1/auth/csrf`, {
        method: "GET",
        credentials: "include",
      });
      const body: ApiEnvelope<{ csrf_token: string }> = await res.json().catch(() => ({}));
      const token = body.data?.csrf_token;
      if (!token) return null;
      csrfToken = token;
      return token;
    } catch {
      return null;
    } finally {
      csrfPromise = null;
    }
  })();

  return csrfPromise;
}

const FETCH_TIMEOUT_MS = 30_000;

function fetchWithTimeout(
  input: string,
  init: RequestInit = {},
): Promise<Response> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), FETCH_TIMEOUT_MS);
  return fetch(input, { ...init, signal: controller.signal }).finally(() =>
    clearTimeout(timeout),
  );
}

async function rawApiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<{ res: Response; body: ApiEnvelope<T> }> {
  const headers = new Headers(options.headers);
  headers.set("Accept", "application/json");
  if (shouldSendJsonBody(options)) {
    headers.set("Content-Type", "application/json");
  }

  if (isMutatingRequest(options.method) && path !== "/api/v1/auth/refresh") {
    const token = await ensureCsrfToken();
    if (token) {
      headers.set("X-CSRF-Token", token);
    }
  }

  const res = await fetchWithTimeout(`${API_URL}${path}`, {
    ...options,
    credentials: "include",
    headers,
  });

  let body: ApiEnvelope<T> = {};
  try {
    body = await res.json();
  } catch {
    body = {};
  }

  return { res, body };
}

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  return request<T>(path, options, false);
}

async function request<T>(
  path: string,
  options: RequestInit,
  isRetry: boolean,
): Promise<T> {
  const { res, body } = await rawApiFetch<T>(path, options);

  if (res.status === 401 && path !== "/api/v1/auth/refresh" && !isRetry) {
    try {
      await refreshAccessToken();
      return request<T>(path, options, true);
    } catch {
      // AuthExpiredError or refresh failure; redirect happens in refreshAccessToken.
      throw new AuthExpiredError();
    }
  }

  if (res.status === 401) {
    redirectToLogin();
    throw new AuthExpiredError();
  }

  if (!res.ok) {
    const err = body.error ?? { code: "UNKNOWN", message: "Request failed" };
    throw new ApiError(res.status, err.code, err.message);
  }

  return body.data as T;
}

async function refreshAccessToken(): Promise<void> {
  if (refreshing) return refreshing;

  refreshing = (async () => {
    const { res } = await rawApiFetch<{ token: string; refresh_token: string }>(
      "/api/v1/auth/refresh",
      {
        method: "POST",
      },
    );

    if (!res.ok) {
      redirectToLogin();
      throw new AuthExpiredError();
    }
  })();

  try {
    return await refreshing;
  } finally {
    refreshing = null;
  }
}

export class AuthExpiredError extends ApiError {
  constructor() {
    super(401, "UNAUTHORIZED", "Session expired. Please sign in again.");
  }
}

async function rawFetch(
  path: string,
  options: RequestInit = {},
): Promise<Response> {
  return fetchWithTimeout(`${API_URL}${path}`, {
    ...options,
    credentials: "include",
  });
}

export async function isAuthenticated(): Promise<boolean> {
  let res = await rawFetch("/api/v1/auth/me", { method: "GET" });

  if (res.status === 401) {
    const refreshRes = await rawFetch("/api/v1/auth/refresh", {
      method: "POST",
    });

    if (refreshRes.ok) {
      res = await rawFetch("/api/v1/auth/me", { method: "GET" });
    } else {
      return false;
    }
  }

  return res.ok;
}

export async function logout(): Promise<void> {
  await apiFetch("/api/v1/auth/logout", { method: "POST" });
}
