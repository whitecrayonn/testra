import { Page, expect } from "@playwright/test";

export async function waitForNetworkIdle(page: Page, timeout = 5000): Promise<void> {
  await page.waitForLoadState("networkidle", { timeout });
}

export async function waitForSelectorStable(page: Page, selector: string, timeout = 5000): Promise<void> {
  await page.waitForSelector(selector, { state: "visible", timeout });
}

export async function waitForElementToDisappear(page: Page, selector: string, timeout = 5000): Promise<void> {
  await page.waitForSelector(selector, { state: "detached", timeout });
}

export async function waitForUrlChange(page: Page, urlPattern: RegExp | string, timeout = 5000): Promise<void> {
  await page.waitForURL(urlPattern, { timeout });
}

export async function waitForText(page: Page, text: string, timeout = 5000): Promise<void> {
  await expect(page.getByText(text).first()).toBeVisible({ timeout });
}

export async function waitForNoSpinners(page: Page, timeout = 5000): Promise<void> {
  const spinnerSelectors = [
    '[data-testid="loading"]',
    '[data-testid="spinner"]',
    ".animate-spin",
    '[aria-busy="true"]',
  ];
  for (const sel of spinnerSelectors) {
    const count = await page.locator(sel).count();
    if (count > 0) {
      await page.locator(sel).last().waitFor({ state: "detached", timeout });
    }
  }
}
