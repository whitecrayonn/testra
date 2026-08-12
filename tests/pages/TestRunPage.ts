import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class TestRunPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.testRuns);
    await this.expectHeading("Test Runs");
  }

  async gotoNew() {
    await this.page.goto(WEB.newTestRun);
    await this.expectHeading("New Test Run");
  }

  async create(name: string, testCaseIds: string[]) {
    await this.gotoNew();
    await this.page.locator("#run-name").fill(name);
    await this.page.locator("#run-case-ids").fill(testCaseIds.join(", "));
    await this.clickByName("Create Run");
    await expect(this.page).toHaveURL(/.*test-runs\/.+/);
  }
}
