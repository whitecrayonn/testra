import { test, expect } from "../../fixtures";
import AxeBuilder from "@axe-core/playwright";
import { LoginPage, RegisterPage, DashboardPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Accessibility - Expanded @a11y @accessibility", () => {
  test("login page has no detectable accessibility violations", async ({ page }) => {
    const login = new LoginPage(page);
    await login.goto();
    const results = await new AxeBuilder({ page }).analyze();
    expect(results.violations).toEqual([]);
  });

  test("register page has no detectable accessibility violations", async ({ page }) => {
    const register = new RegisterPage(page);
    await register.goto();
    const results = await new AxeBuilder({ page }).analyze();
    expect(results.violations).toEqual([]);
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
    const results = await new AxeBuilder({ page: authPage })
      .disableRules(["color-contrast"])
      .analyze();
    expect(results.violations).toEqual([]);
  });

  test("login page has proper labels for form inputs", async ({ page }) => {
    const login = new LoginPage(page);
    await login.goto();
    await expect(page.getByLabel("Email")).toBeVisible();
    await expect(page.getByLabel("Password")).toBeVisible();
  });

  test("register page has proper labels for form inputs", async ({ page }) => {
    const register = new RegisterPage(page);
    await register.goto();
    await expect(page.getByLabel(/email/i)).toBeVisible();
    await expect(page.getByLabel(/password/i)).toBeVisible();
  });

  test("login page has proper heading structure", async ({ page }) => {
    const login = new LoginPage(page);
    await login.goto();
    await expect(page.getByRole("heading", { name: /sign in to testra/i })).toBeVisible();
  });

  test("dashboard has proper heading structure", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const dashboard = new DashboardPage(authPage);
    await dashboard.goto();
    await expect(authPage.getByRole("heading", { name: "Dashboard" })).toBeVisible();
  });

  test("login page error messages have role alert", async ({ page }) => {
    const login = new LoginPage(page);
    await login.goto();
    await page.getByLabel("Email").fill("invalid@example.com");
    await page.getByLabel("Password").fill("wrongpassword");
    await page.getByRole("button", { name: /sign in/i }).click();
    await expect(page.locator('p[role="alert"]')).toBeVisible({ timeout: 10000 });
  });

  test("login page is keyboard navigable", async ({ page }) => {
    const login = new LoginPage(page);
    await login.goto();
    await page.keyboard.press("Tab");
    await page.keyboard.type("test@example.com");
    await page.keyboard.press("Tab");
    await page.keyboard.type("password");
    await page.keyboard.press("Tab");
    await page.keyboard.press("Enter");
    await page.waitForURL(/\/dashboard|\/login/, { timeout: 10000 });
  });
});
