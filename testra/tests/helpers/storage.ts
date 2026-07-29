import { Page } from "@playwright/test";

export interface WorkspaceContext {
  workspaceId: string;
  workspaceName: string;
  organizationId: string;
  projectId?: string;
  projectName?: string;
}

export async function setWorkspaceContext(page: Page, ctx: WorkspaceContext): Promise<void> {
  if (page.url() === "about:blank") {
    await page.goto("/create-workspace", { waitUntil: "domcontentloaded" });
  }
  await page.evaluate((c: WorkspaceContext) => {
    localStorage.setItem("testra_workspace_id", c.workspaceId);
    localStorage.setItem("testra_workspace_name", c.workspaceName);
    localStorage.setItem("testra_organization_id", c.organizationId);
    if (c.projectId) {
      localStorage.setItem("testra_project_id", c.projectId);
      if (c.projectName) localStorage.setItem("testra_project_name", c.projectName);
    }
  }, ctx);
}

export async function clearWorkspaceContext(page: Page): Promise<void> {
  await page.evaluate(() => {
    localStorage.removeItem("testra_workspace_id");
    localStorage.removeItem("testra_workspace_name");
    localStorage.removeItem("testra_organization_id");
    localStorage.removeItem("testra_project_id");
    localStorage.removeItem("testra_project_name");
  });
}

export async function getWorkspaceContext(page: Page): Promise<Partial<WorkspaceContext>> {
  return page.evaluate(() => ({
    workspaceId: localStorage.getItem("testra_workspace_id") || undefined,
    workspaceName: localStorage.getItem("testra_workspace_name") || undefined,
    organizationId: localStorage.getItem("testra_organization_id") || undefined,
    projectId: localStorage.getItem("testra_project_id") || undefined,
    projectName: localStorage.getItem("testra_project_name") || undefined,
  }));
}
