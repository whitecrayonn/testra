import { test, expect } from "../../fixtures";
import { LoginPage } from "../../pages";
import { API } from "../../constants/routes";

test.describe("Smoke tests", { tag: "@smoke" }, () => {
  test("public health endpoint returns success", async ({ api }) => {
    const res = await api.request.get(`${api.baseURL}${API.health}`);
    expect(res.ok()).toBeTruthy();
  });

  test("login page loads", async ({ page }) => {
    const login = new LoginPage(page);
    await login.goto();
    await expect(page).toHaveURL(/.*\/login/);
  });

  test("authenticated user can access dashboard", async ({ authPage }) => {
    await authPage.goto("/dashboard");
    await expect(authPage.getByRole("heading", { name: "Dashboard" })).toBeVisible();
  });
});
