import createClient from "openapi-fetch";
import type { paths } from "./openapi";

export type { paths };

export function createTestraClient(baseUrl: string) {
  return createClient<paths>({ baseUrl });
}
