import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class DashboardPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.dashboard);
    await this.expectHeading("Dashboard");
  }

  async expectLoaded() {
    await this.expectHeading("Dashboard");
    await expect(this.page.getByText("Workspace context")).toBeVisible();
  }

  async navigateToTestCases() {
    await this.page.getByRole("link", { name: "Test Cases", exact: true }).click();
    await expect(this.page).toHaveURL(/.*test-cases/);
  }

  async navigateToTestRuns() {
    await this.page.getByRole("link", { name: "Runs", exact: true }).click();
    await expect(this.page).toHaveURL(/.*test-runs/);
  }

  async navigateToProjects() {
    await this.page.getByRole("link", { name: "Projects", exact: true }).click();
    await expect(this.page).toHaveURL(/.*projects/);
  }

  async navigateToDefects() {
    await this.page.goto(WEB.defects);
    await this.expectHeading("Defects");
  }

  async navigateToAutomation() {
    await this.page.goto(WEB.automation);
    await this.expectHeading("Automation Hub");
  }

  async navigateToApiTests() {
    await this.page.goto(WEB.apiTests);
    await this.expectHeading("API Testing");
  }

  async navigateToAnalytics() {
    await this.page.goto(WEB.analytics);
    await this.expectHeading("Executive");
  }

  async navigateToNotifications() {
    await this.page.goto(WEB.notifications);
    await this.expectHeading("Notifications");
  }

  async navigateToSettings() {
    await this.page.goto(WEB.settings);
    await this.expectHeading("Settings");
  }
}
