import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class IntegrationPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.integrations);
    await this.expectHeading("Integrations");
  }

  async openCreate() {
    await this.page.getByRole("button", { name: /add integration|new integration|connect/i }).click();
  }

  async fillName(name: string) {
    await this.fillByLabel(/integration name|name/i, name);
  }

  async selectProvider(provider: string) {
    await this.page.locator('label:has-text("Provider") + select, label:has-text("Provider") select').selectOption(provider);
  }

  async fillWebhookUrl(url: string) {
    await this.fillByLabel(/webhook url|url/i, url);
  }

  async submit() {
    await this.clickByName(/save|create|connect/i);
  }

  async create(name: string, provider: string, webhookUrl?: string) {
    await this.goto();
    await this.openCreate();
    await this.fillName(name);
    await this.selectProvider(provider);
    if (webhookUrl) await this.fillWebhookUrl(webhookUrl);
    await this.submit();
    await expect(this.page.getByText(name)).toBeVisible();
  }

  async toggleIntegration(name: string) {
    const row = this.page.locator(`[data-testid="integration-row"]:has-text("${name}")`).or(
      this.page.locator(`tr:has-text("${name}")`).or(this.page.locator(`div:has-text("${name}")`)),
    ).first();
    await row.getByRole("switch").or(row.getByRole("button", { name: /enable|disable|toggle/i })).click();
  }

  async expectIntegrationVisible(name: string) {
    await expect(this.page.getByText(name).first()).toBeVisible();
  }
}
