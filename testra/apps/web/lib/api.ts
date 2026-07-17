const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

const TOKEN_KEY = "testra_token";
const REFRESH_TOKEN_KEY = "testra_refresh_token";

let refreshing: Promise<{ token: string; refresh_token: string }> | null = null;

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

async function rawApiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<{ res: Response; body: ApiEnvelope<T> }> {
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
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
  const token = getToken();

  const attempt = async (): Promise<T> => {
    const { res, body } = await rawApiFetch<T>(path, {
      ...options,
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.headers,
      },
    });

    if (res.status === 401 && path !== "/api/v1/auth/refresh") {
      const newToken = await refreshAccessToken();
      return apiFetch(path, {
        ...options,
        headers: {
          ...options.headers,
          Authorization: `Bearer ${newToken.token}`,
        },
      });
    }

    if (!res.ok) {
      const err = body.error ?? { code: "UNKNOWN", message: "Request failed" };
      throw new ApiError(res.status, err.code, err.message);
    }

    return body.data as T;
  };

  return attempt();
}

async function refreshAccessToken(): Promise<{ token: string; refresh_token: string }> {
  if (refreshing) return refreshing;

  const refreshToken = getRefreshToken();
  if (!refreshToken) {
    throw new AuthExpiredError();
  }

  refreshing = (async () => {
    const { res, body } = await rawApiFetch<{ token: string; refresh_token: string }>(
      "/api/v1/auth/refresh",
      {
        method: "POST",
        body: JSON.stringify({ refresh_token: refreshToken }),
      },
    );

    if (!res.ok || !body.data) {
      clearAuth();
      if (typeof window !== "undefined") {
        window.location.href = "/login";
      }
      throw new AuthExpiredError();
    }

    setAuth(body.data.token, body.data.refresh_token);
    return body.data;
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

export function setAuth(token: string, refreshToken: string) {
  if (typeof window !== "undefined") {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
  }
}

export function clearAuth() {
  if (typeof window !== "undefined") {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
  }
}

export function setToken(token: string) {
  if (typeof window !== "undefined") {
    localStorage.setItem(TOKEN_KEY, token);
  }
}

export function clearToken() {
  if (typeof window !== "undefined") {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
  }
}

export function getToken(): string | null {
  if (typeof window !== "undefined") {
    return localStorage.getItem(TOKEN_KEY);
  }
  return null;
}

export function getRefreshToken(): string | null {
  if (typeof window !== "undefined") {
    return localStorage.getItem(REFRESH_TOKEN_KEY);
  }
  return null;
}

export function isAuthenticated(): boolean {
  return getToken() !== null;
}
