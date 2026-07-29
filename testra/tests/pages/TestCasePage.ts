import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class TestCasePage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.testCases);
    await this.expectHeading("Test Cases");
  }

  async gotoNew() {
    await this.page.goto(WEB.newTestCase);
    await this.expectHeading("New Test Case");
  }

  async create(title: string, description?: string, steps?: string[]) {
    await this.gotoNew();
    await this.fillByLabel("Title", title);
    if (description) await this.fillByLabel("Description", description);
    if (steps && steps.length > 0) {
      for (const step of steps) {
        await this.page.getByRole("button", { name: /add step/i }).click();
        await this.page.getByLabel(/action/i).last().fill(step);
        await this.page.getByLabel(/expected/i).last().fill(`Expected: ${step}`);
      }
    }
    await this.clickByName("Create test case");
    await expect(this.page).toHaveURL(/.*test-cases/);
  }

  async search(query: string) {
    await this.goto();
    const searchInput = this.page.getByPlaceholder(/search/i);
    await searchInput.fill(query);
    await searchInput.press("Enter");
  }

  async expectTestCase(title: string) {
    await expect(this.page.getByText(title)).toBeVisible();
  }
}
