import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class TestPlanPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.testPlans);
    await this.expectHeading("Test Plans");
  }

  async gotoNew() {
    await this.page.goto(WEB.newTestPlan);
    await this.expectHeading("New Test Plan");
  }

  async create(name: string, testCaseIds: string[]) {
    await this.gotoNew();
    await this.fillByLabel("Name", name);
    await this.fillByLabel("Description", "Automated test plan");
    await this.fillByLabel("Test Case IDs", testCaseIds.join(", "));
    await this.clickByName("Create plan");
    await expect(this.page).toHaveURL(/.*test-plans/);
  }
}
