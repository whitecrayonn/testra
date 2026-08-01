import { Page, Locator, expect } from "@playwright/test";

export class BasePage {
  constructor(protected page: Page) {}

  async goto(path: string, options?: { waitFor?: string }) {
    await this.page.goto(path);
    if (options?.waitFor) {
      await this.page.waitForSelector(options.waitFor, { state: "visible" });
    }
  }

  async expectHeading(title: string) {
    await expect(this.page.getByRole("heading", { name: title, exact: false })).toBeVisible({ timeout: 30000 });
  }

  async expectUrl(path: string) {
    await expect(this.page).toHaveURL(new RegExp(`.*${path}`));
  }

  getByLabel(label: string | RegExp): Locator {
    return this.page.getByLabel(label);
  }

  getByRole(role: string, name?: string | RegExp): Locator {
    return this.page.getByRole(role as any, name ? { name } : undefined);
  }

  getByText(text: string | RegExp): Locator {
    return this.page.getByText(text);
  }

  getByTestId(testId: string): Locator {
    return this.page.getByTestId(testId);
  }

  async fillByLabel(label: string | RegExp, value: string) {
    await this.getByLabel(label).fill(value);
  }

  async clickByName(name: string | RegExp) {
    await this.page.getByRole("button", { name }).click();
  }

  async waitForLoading() {
    await this.page.waitForLoadState("networkidle");
  }

  async hasError(message?: string | RegExp) {
    const alert = this.page
      .locator('[role="alert"]:not(#__next-route-announcer__)')
      .filter({ hasText: message || /error|failed|invalid/i });
    await expect(alert).toBeVisible();
  }
}
