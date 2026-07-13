const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

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

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token =
    typeof window !== "undefined" ? localStorage.getItem("testra_token") : null;

  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  });

  const body: ApiEnvelope<T> = await res.json();

  if (!res.ok) {
    const err = body.error ?? { code: "UNKNOWN", message: "Request failed" };
    throw new ApiError(res.status, err.code, err.message);
  }

  return body.data as T;
}

export function setToken(token: string) {
  if (typeof window !== "undefined") {
    localStorage.setItem("testra_token", token);
  }
}

export function clearToken() {
  if (typeof window !== "undefined") {
    localStorage.removeItem("testra_token");
  }
}

export function getToken(): string | null {
  if (typeof window !== "undefined") {
    return localStorage.getItem("testra_token");
  }
  return null;
}
