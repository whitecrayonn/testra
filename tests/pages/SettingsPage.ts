import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class SettingsPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.settings);
    await this.expectHeading("Settings");
  }

  async gotoMembers() {
    await this.page.goto(WEB.members);
    await this.expectHeading("Members");
  }

  async gotoIntegrations() {
    await this.page.goto(WEB.integrations);
    await this.expectHeading("Integrations");
  }

  async inviteMember(email: string, role: string) {
    await this.gotoMembers();
    await this.fillByLabel(/email/i, email);
    await this.page.locator('label:has-text("Role") + select, label:has-text("Role") select').selectOption(role);
    await this.clickByName(/invite|send invite/i);
  }

  async expectMember(email: string) {
    await expect(this.page.getByText(email).first()).toBeVisible();
  }

  async updateWorkspaceName(name: string) {
    await this.goto();
    await this.fillByLabel(/workspace name/i, name);
    await this.clickByName(/save|update/i);
  }
}
