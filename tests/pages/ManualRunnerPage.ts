import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class ManualRunnerPage extends BasePage {
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

  async openRun(runId: string) {
    await this.page.goto(`/dashboard/test-runs/${runId}`);
  }

  async setStepStatus(stepIndex: number, status: "passed" | "failed" | "skipped" | "blocked") {
    const step = this.page.locator(`[data-testid="run-item-${stepIndex}"]`).or(
      this.page.locator(`tr, div`).filter({ hasText: `Step ${stepIndex + 1}` }),
    ).first();
    const button = step.getByRole("button", { name: new RegExp(status, "i") });
    if (await button.count() > 0) {
      await button.click();
    } else {
      await step.locator("select").selectOption(status);
    }
  }

  async expectRunStatus(status: string) {
    await expect(this.page.getByText(new RegExp(status, "i")).first()).toBeVisible();
  }

  async addComment(comment: string) {
    await this.page.getByPlaceholder(/comment|note/i).fill(comment);
    await this.clickByName(/add comment|submit comment/i);
  }
}
