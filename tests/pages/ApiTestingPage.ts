import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class ApiTestingPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.apiTests);
    await this.expectHeading("API Testing");
  }

  async createCollection(name: string) {
    await this.goto();
    const newCollectionInput = this.page.getByPlaceholder("New collection").first();
    await newCollectionInput.fill(name);
    await newCollectionInput.press("Enter");
    await expect(this.page.getByText(name).first()).toBeAttached();
  }

  async createEnvironment(name: string) {
    await this.page.getByRole("button", { name: "Environments" }).click();
    await this.page.getByPlaceholder("New environment").fill(name);
    await this.page.getByRole("button", { name: "Add" }).click();
    await expect(this.page.getByText(name)).toBeVisible();
  }

  async createRequest(url: string) {
    await this.page.getByRole("button", { name: "Requests" }).click();
    await this.page.locator('input[placeholder*="Request name"]').fill("Sample Request");
    await this.page.locator('input[placeholder*="https://"]').fill(url);
    await this.page.getByRole("button", { name: /send/i }).click();
    await expect(this.page.locator("text=Response").or(this.page.getByText("200"))).toBeVisible();
  }
}
