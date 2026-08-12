import { APIRequestContext, expect } from "@playwright/test";

export interface PaginatedResponse<T> {
  data: T[];
  meta: {
    next_cursor?: string;
    has_more: boolean;
  };
}

export async function fetchAllPages<T>(
  request: APIRequestContext,
  baseUrl: string,
  params: Record<string, string> = {},
  maxPages = 10,
): Promise<T[]> {
  const allItems: T[] = [];
  let cursor = "";
  let pageCount = 0;

  while (pageCount < maxPages) {
    const queryParams = new URLSearchParams(params);
    if (cursor) {
      queryParams.set("cursor", cursor);
    }
    const url = `${baseUrl}?${queryParams.toString()}`;
    const response = await request.get(url);
    expect(response.ok()).toBeTruthy();
    const body = await response.json() as PaginatedResponse<T>;
    allItems.push(...body.data);
    if (!body.meta.has_more || !body.meta.next_cursor) {
      break;
    }
    cursor = body.meta.next_cursor;
    pageCount++;
  }

  return allItems;
}

export async function fetchPage<T>(
  request: APIRequestContext,
  baseUrl: string,
  params: Record<string, string> = {},
  cursor?: string,
  limit?: number,
): Promise<PaginatedResponse<T>> {
  const queryParams = new URLSearchParams(params);
  if (cursor) {
    queryParams.set("cursor", cursor);
  }
  if (limit) {
    queryParams.set("limit", String(limit));
  }
  const url = `${baseUrl}?${queryParams.toString()}`;
  const response = await request.get(url);
  expect(response.ok()).toBeTruthy();
  return await response.json() as PaginatedResponse<T>;
}

export function decodeCursor(cursor: string): string {
  const decoded = Buffer.from(cursor, "base64").toString("utf-8");
  const parsed = JSON.parse(decoded);
  return parsed.id;
}
