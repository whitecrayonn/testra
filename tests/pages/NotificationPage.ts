import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class NotificationPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.notifications);
    await this.expectHeading("Notifications");
  }

  async switchToUnread() {
    await this.page.getByRole("button", { name: "Unread" }).click();
  }

  async markFirstRead() {
    await this.page.getByRole("button", { name: "Mark read" }).first().click();
    await expect(this.page.getByRole("button", { name: "Mark unread" }).first()).toBeVisible();
  }
}
