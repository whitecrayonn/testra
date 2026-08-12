import { Page, expect } from "@playwright/test";

export async function expectElementVisible(page: Page, selector: string, timeout = 5000): Promise<void> {
  await expect(page.locator(selector).first()).toBeVisible({ timeout });
}

export async function expectElementNotVisible(page: Page, selector: string, timeout = 5000): Promise<void> {
  await expect(page.locator(selector).first()).not.toBeVisible({ timeout });
}

export async function expectTextVisible(page: Page, text: string, timeout = 5000): Promise<void> {
  await expect(page.getByText(text).first()).toBeVisible({ timeout });
}

export async function expectUrlContains(page: Page, fragment: string): Promise<void> {
  await expect(page).toHaveURL(new RegExp(fragment));
}

export async function expectTitleContains(page: Page, text: string): Promise<void> {
  await expect(page).toHaveTitle(new RegExp(text, "i"));
}

export async function expectCount(page: Page, selector: string, count: number): Promise<void> {
  await expect(page.locator(selector)).toHaveCount(count);
}

export async function expectGreaterThan(page: Page, selector: string, minCount: number): Promise<void> {
  const count = await page.locator(selector).count();
  expect(count).toBeGreaterThan(minCount);
}

export async function expectEmptyList(page: Page, selector: string): Promise<void> {
  await expect(page.locator(selector)).toHaveCount(0);
}

export async function expectButtonEnabled(page: Page, name: string | RegExp): Promise<void> {
  await expect(page.getByRole("button", { name })).toBeEnabled();
}

export async function expectButtonDisabled(page: Page, name: string | RegExp): Promise<void> {
  await expect(page.getByRole("button", { name })).toBeDisabled();
}

export async function expectInputValue(page: Page, label: string | RegExp, value: string): Promise<void> {
  const input = page.getByLabel(label);
  await expect(input).toHaveValue(value);
}
