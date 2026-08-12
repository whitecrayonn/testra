import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class ProjectPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await this.page.goto(WEB.projects);
    await this.expectHeading("Projects");
  }

  async openCreate() {
    await this.page.getByRole("button", { name: /new project|create project/i }).click();
  }

  async fillName(name: string) {
    await this.fillByLabel("Project name", name);
  }

  async fillKey(key: string) {
    await this.fillByLabel("Project key", key);
  }

  async fillDescription(description: string) {
    await this.fillByLabel("Description", description);
  }

  async submit() {
    await this.clickByName(/create project|save/i);
  }

  async create(name: string, key: string, description?: string) {
    await this.goto();
    await this.openCreate();
    await this.fillName(name);
    await this.fillKey(key);
    await this.submit();
    await expect(this.page.getByText(name)).toBeVisible();
  }

  async selectProject(name: string) {
    await this.page.getByRole("link", { name }).click();
    await expect(this.page.getByText("Project selected")).toBeVisible();
  }
}
