import { test, expect } from "../../fixtures";
import { LoginPage, DashboardPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Visual Regression @visual", () => {
  test("login page screenshot", async ({ page }) => {
    const login = new LoginPage(page);
    await login.goto();
    await expect(page).toHaveScreenshot("login.png");
  });

  test("dashboard screenshot", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const dashboard = new DashboardPage(authPage);
    await dashboard.goto();
    await expect(authPage).toHaveScreenshot("dashboard.png");
  });
});
