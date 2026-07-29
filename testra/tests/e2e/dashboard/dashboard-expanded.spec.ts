import { test, expect } from "../../fixtures";
import { DashboardPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Dashboard - Expanded @dashboard @regression", () => {
  test("dashboard displays workspace context", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const dashboard = new DashboardPage(authPage);
    await dashboard.goto();
    await expect(authPage.getByText(workspace.name)).toBeVisible();
  });

  test("dashboard shows project badge when project is set", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const dashboard = new DashboardPage(authPage);
    await dashboard.goto();
    await expect(authPage.getByText(project.name).first()).toBeVisible();
  });

  test("dashboard shows no project warning when project is not set", async ({ authPage, workspace }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: "",
      projectName: "",
    });

    const dashboard = new DashboardPage(authPage);
    await dashboard.goto();
    await expect(authPage.getByText(/no project/i)).toBeVisible();
  });

  test("dashboard has quick action links", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const dashboard = new DashboardPage(authPage);
    await dashboard.goto();
    await expect(authPage.getByRole("link", { name: /test runs/i }).first()).toBeVisible();
    await expect(authPage.getByRole("link", { name: /test cases/i }).first()).toBeVisible();
    await expect(authPage.getByRole("link", { name: /projects/i }).first()).toBeVisible();
  });

  test("dashboard quick action new test case navigates correctly", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const dashboard = new DashboardPage(authPage);
    await dashboard.goto();
    await authPage.getByRole("link", { name: /new test case/i }).click();
    await expect(authPage).toHaveURL(/.*test-cases\/new.*/);
  });

  test("dashboard quick action new test run navigates correctly", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const dashboard = new DashboardPage(authPage);
    await dashboard.goto();
    await authPage.getByRole("link", { name: /new test run/i }).click();
    await expect(authPage).toHaveURL(/.*test-runs\/new.*/);
  });

  test("dashboard redirects unauthenticated user to login", async ({ page }) => {
    await page.goto("/dashboard");
    await expect(page).toHaveURL(/.*login.*/);
  });

  test("dashboard analytics section loads when workspace is set", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const dashboard = new DashboardPage(authPage);
    await dashboard.goto();
    await expect(authPage.getByText("Workspace context")).toBeVisible();
  });
});
