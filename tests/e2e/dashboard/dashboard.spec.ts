import { test } from "../../fixtures";
import { DashboardPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Dashboard @dashboard", () => {
  test("authenticated user sees dashboard workspace context", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const dashboard = new DashboardPage(authPage);
    await dashboard.goto();
    await dashboard.expectLoaded();
  });

  test("dashboard navigation links work", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const dashboard = new DashboardPage(authPage);
    await dashboard.goto();
    await dashboard.navigateToTestCases();
    await dashboard.navigateToTestRuns();
  });
});
