import { Page, expect } from "@playwright/test";
import { BasePage } from "./BasePage";
import { WEB } from "../constants/routes";

export class WorkspacePage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async gotoCreate() {
    await this.page.goto(WEB.createWorkspace);
    await this.expectHeading("Create your workspace");
  }

  async fillWorkspaceName(name: string) {
    await this.fillByLabel("Workspace name", name);
  }

  async fillSlug(slug: string) {
    await this.fillByLabel("Slug", slug);
  }

  async submit() {
    await this.clickByName("Create workspace");
  }

  async create(name: string, slug?: string) {
    await this.gotoCreate();
    await this.fillWorkspaceName(name);
    if (slug) await this.fillSlug(slug);
    await this.submit();
    await expect(this.page).toHaveURL(/.*\/dashboard/);
  }
}
