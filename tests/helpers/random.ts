export function uniqueId(prefix = "testra"): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

export function uniqueEmail(prefix = "testra"): string {
  return `${prefix}+${Date.now()}-${Math.random().toString(36).slice(2, 6)}@example.com`;
}

export function slugify(name: string): string {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, "")
    .replace(/\s+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-+|-+$/g, "");
}

export function generatePassword(): string {
  return `P@ssw0rd!${Date.now()}${Math.random().toString(36).slice(2, 6)}`;
}

export function todayIso(): string {
  return new Date().toISOString();
}
