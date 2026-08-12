const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "";

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

// Extends RequestInit with `silent`, which excludes a request from the
// global pending-request counter below — used for background polling
// (e.g. unread-count) that shouldn't trigger a visible loading indicator.
export interface ApiRequestInit extends RequestInit {
  silent?: boolean;
}

type PendingListener = (count: number) => void;
const pendingListeners = new Set<PendingListener>();
let pendingRequestCount = 0;

function notifyPending() {
  pendingListeners.forEach((listener) => listener(pendingRequestCount));
}

/** Subscribes to the number of in-flight (non-silent) API requests. Used to drive a global loading indicator. */
export function subscribeToPendingRequests(listener: PendingListener): () => void {
  pendingListeners.add(listener);
  listener(pendingRequestCount);
  return () => {
    pendingListeners.delete(listener);
  };
}

function fetchWithTimeout(
  input: string,
  init: ApiRequestInit = {},
): Promise<Response> {
  const { silent, ...requestInit } = init;
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), FETCH_TIMEOUT_MS);
  if (!silent) {
    pendingRequestCount++;
    notifyPending();
  }
  return fetch(input, { ...requestInit, signal: controller.signal }).finally(() => {
    clearTimeout(timeout);
    if (!silent) {
      pendingRequestCount--;
      notifyPending();
    }
  });
}

async function rawApiFetch<T>(
  path: string,
  options: ApiRequestInit = {},
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
  options: ApiRequestInit = {},
): Promise<T> {
  return request<T>(path, options, false);
}

export interface PaginatedResult<T> {
  data: T[];
  meta: Record<string, unknown>;
}

export async function apiFetchWithMeta<T>(
  path: string,
  options: ApiRequestInit = {},
): Promise<PaginatedResult<T>> {
  const { res, body } = await rawApiFetch<T[]>(path, options);

  if (res.status === 401 && path !== "/api/v1/auth/refresh" && path !== "/api/v1/auth/login") {
    try {
      await refreshAccessToken();
      const retry = await rawApiFetch<T[]>(path, options);
      return {
        data: retry.body.data ?? [],
        meta: (retry.body.meta as Record<string, unknown>) ?? {},
      };
    } catch {
      throw new AuthExpiredError();
    }
  }

  if (res.status === 401 && path !== "/api/v1/auth/login" && path !== "/api/v1/auth/register") {
    redirectToLogin();
    throw new AuthExpiredError();
  }

  if (!res.ok) {
    const err = body.error ?? { code: "UNKNOWN", message: "Request failed" };
    throw new ApiError(res.status, err.code, err.message);
  }

  return {
    data: body.data ?? [],
    meta: (body.meta as Record<string, unknown>) ?? {},
  };
}

async function request<T>(
  path: string,
  options: ApiRequestInit,
  isRetry: boolean,
): Promise<T> {
  const { res, body } = await rawApiFetch<T>(path, options);

  if (res.status === 401 && path !== "/api/v1/auth/refresh" && path !== "/api/v1/auth/login" && !isRetry) {
    try {
      await refreshAccessToken();
      return request<T>(path, options, true);
    } catch {
      // AuthExpiredError or refresh failure; redirect happens in refreshAccessToken.
      throw new AuthExpiredError();
    }
  }

  if (res.status === 401 && path !== "/api/v1/auth/login" && path !== "/api/v1/auth/register") {
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
