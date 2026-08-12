import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class LoginPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.login);
    await this.expectHeading("Sign in to Testra");
  }

  async fillEmail(email: string) {
    await this.fillByLabel("Email", email);
  }

  async fillPassword(password: string) {
    await this.fillByLabel("Password", password);
  }

  async submit() {
    await this.clickByName("Sign in");
  }

  async login(email: string, password: string) {
    await this.goto();
    await this.fillEmail(email);
    await this.fillPassword(password);
    await this.submit();
    await expect(this.page).toHaveURL(/.*\/dashboard/);
  }

  async expectMfaPrompt() {
    await expect(this.page.getByLabel("MFA Code")).toBeVisible();
  }

  async submitMfa(code: string) {
    await this.fillByLabel("MFA Code", code);
    await this.clickByName("Sign in");
  }
}
