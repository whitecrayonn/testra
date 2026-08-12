import { test, expect } from "../../fixtures";
import { WorkspacePage, LoginPage } from "../../pages";
import { uniqueId } from "../../helpers/random";

test.describe("Workspace Onboarding @onboarding @smoke", () => {
  test("new user can register and create workspace through onboarding", async ({ page, user }) => {
    const login = new LoginPage(page);
    await login.login(user.user.email, user.password);

    await page.waitForURL(/.*create-workspace/);
    const workspace = new WorkspacePage(page);
    const slug = uniqueId("ws");
    await workspace.create(`E2E Workspace ${slug}`, slug);

    await expect(page).toHaveURL(/.*\/dashboard/);
    await expect(page.getByRole("heading", { name: "Dashboard" })).toBeVisible();
  });

  test("onboarding flow creates organization and workspace", async ({ authPage, api, user }) => {
    await api.login({ email: user.user.email, password: user.password });
    const workspace = new WorkspacePage(authPage);
    const slug = uniqueId("ws");
    await workspace.create(`E2E Workspace ${slug}`, slug);

    const orgs = await api.listOrganizations();
    expect(orgs.length).toBeGreaterThan(0);
    const workspaces = await api.listWorkspaces(orgs[0].id);
    expect(workspaces.length).toBeGreaterThan(0);
  });
});
