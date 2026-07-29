import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class AutomationPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.automation);
    await this.expectHeading("Automation Hub");
  }

  async openCreate() {
    await this.page.getByRole("button", { name: /new project/i }).click();
  }

  async create(name: string, framework = "junit") {
    await this.goto();
    await this.openCreate();
    const nameInput = this.page.locator('input[placeholder*="Project name"], input[placeholder*="project name"]').first();
    await nameInput.fill(name);
    await this.page.locator("select").first().selectOption(framework);
    await this.clickByName("Create");
    await expect(this.page.getByText(name)).toBeVisible();
  }
}
