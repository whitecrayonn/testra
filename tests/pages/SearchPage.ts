import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";

export class SearchPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async search(query: string) {
    await this.page.waitForTimeout(1000);
    await this.page.evaluate(() => document.dispatchEvent(new CustomEvent("open-global-search")));
    const searchInput = this.page.getByPlaceholder(/search/i).or(this.page.locator('input[type="search"]')).first();
    await searchInput.waitFor({ state: "visible" });
    await searchInput.fill(query);
    await searchInput.press("Enter");
  }

  async expectResult(title: string) {
    await expect(this.page.getByText(title).first()).toBeVisible();
  }
}
