import { test, expect } from "../../fixtures";
import { LoginPage, RegisterPage, DashboardPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Visual Regression - Expanded @visual @regression", () => {
  test("login page screenshot", async ({ page }) => {
    const login = new LoginPage(page);
    await login.goto();
    await expect(page).toHaveScreenshot("login.png", { maxDiffPixelRatio: 0.1 });
  });

  test("register page screenshot", async ({ page }) => {
    const register = new RegisterPage(page);
    await register.goto();
    await expect(page).toHaveScreenshot("register.png", { maxDiffPixelRatio: 0.1 });
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
    await expect(authPage).toHaveScreenshot("dashboard.png", { maxDiffPixelRatio: 0.1 });
  });

  test("dashboard without project context screenshot", async ({ authPage, workspace }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: "",
      projectName: "",
    });

    const dashboard = new DashboardPage(authPage);
    await dashboard.goto();
    await expect(authPage).toHaveScreenshot("dashboard-no-project.png", { maxDiffPixelRatio: 0.1 });
  });
});
