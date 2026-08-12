import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function slugify(s: string): string {
  const result = s
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, "")
    .replace(/[\s_-]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return result || "workspace";
}

/** Derives a project key from a name (e.g. "Web App QA" -> "WEBAPPQA"):
 * uppercase, alphanumeric only, max 10 chars, never starts with a digit,
 * never shorter than 2 chars. Falls back to "PROJECT" if nothing usable
 * remains. Shared by the Projects "New Project" form and the "Create
 * workspace" flow's auto-created first project, so both produce the same
 * key for the same name. */
export function generateProjectKey(name: string): string {
  const cleaned = name.toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, 10);
  if (!cleaned) return "PROJECT";
  if (/^[0-9]/.test(cleaned)) {
    return "P" + cleaned.slice(0, 9);
  }
  if (cleaned.length < 2) return cleaned + "1";
  return cleaned;
}
