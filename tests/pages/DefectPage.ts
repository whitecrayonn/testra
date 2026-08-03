import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class DefectPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.defects);
    await this.expectHeading("Defects");
  }

  async openCreate() {
    await this.page.getByRole("button", { name: /new defect/i }).click();
  }

  async create(title: string, description?: string, severity = "medium", priority = "medium") {
    await this.goto();
    await this.openCreate();
    await this.fillByLabel("Title", title);
    if (description) await this.fillByLabel("Description", description);
    await this.page.locator('label:has-text("Severity") + select, label:has-text("Severity") select').selectOption(severity);
    await this.page.locator('label:has-text("Priority") + select, label:has-text("Priority") select').selectOption(priority);
    await this.clickByName("Create defect");
    await expect(this.page.getByText(title)).toBeVisible();
  }
}
