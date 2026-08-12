import { test, expect } from "../../fixtures";
import AxeBuilder from "@axe-core/playwright";
import { LoginPage, DashboardPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Accessibility @a11y", () => {
  test("login page has no detectable accessibility violations", async ({ page }) => {
    const login = new LoginPage(page);
    await login.goto();
    const accessibilityScanResults = await new AxeBuilder({ page }).analyze();
    expect(accessibilityScanResults.violations).toEqual([]);
  });

  test("dashboard has no critical accessibility violations", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const dashboard = new DashboardPage(authPage);
    await dashboard.goto();
    const accessibilityScanResults = await new AxeBuilder({ page: authPage })
      .disableRules(["color-contrast"])
      .analyze();
    expect(accessibilityScanResults.violations).toEqual([]);
  });
});
