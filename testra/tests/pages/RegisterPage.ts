import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class RegisterPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.register);
    await this.expectHeading("Create your account");
  }

  async fillName(name: string) {
    await this.fillByLabel("Full name", name);
  }

  async fillEmail(email: string) {
    await this.fillByLabel("Email", email);
  }

  async fillPassword(password: string) {
    await this.fillByLabel("Password", password);
  }

  async submit() {
    await this.clickByName("Create account");
  }

  async register(name: string, email: string, password: string) {
    await this.goto();
    await this.fillName(name);
    await this.fillEmail(email);
    await this.fillPassword(password);
    await this.submit();
    await expect(this.page).toHaveURL(/.*create-workspace/);
  }
}
