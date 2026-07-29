import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class AnalyticsPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.analytics);
    await this.expectHeading("Executive Dashboard");
  }

  async expectSummaryVisible() {
    await expect(this.page.getByText(/summary|metrics|test cases|runs/i).first()).toBeVisible();
  }
}
