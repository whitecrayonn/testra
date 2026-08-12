import { test, expect } from "../../fixtures";
import { LoginPage, RegisterPage } from "../../pages";

test.describe("Authentication @auth @smoke", () => {
  test("user can register and log in", async ({ page, user }) => {
    const login = new LoginPage(page);
    await login.login(user.user.email, user.password);

    await expect(page).toHaveURL(/.*\/dashboard/);
    await expect(page.getByRole("heading", { name: "Dashboard" })).toBeVisible();
  });

  test("login with invalid credentials shows error", async ({ page }) => {
    const login = new LoginPage(page);
    await login.goto();
    await login.fillEmail("notfound@example.com");
    await login.fillPassword("wrong-password-123");
    await login.submit();
    await login.hasError(/invalid|failed|not found/i);
  });

  test("unauthenticated user is redirected to login", async ({ page }) => {
    await page.goto("/dashboard");
    await expect(page).toHaveURL(/.*\/login/);
  });

  test("register enforces password minimum length", async ({ page }) => {
    const register = new RegisterPage(page);
    await register.goto();
    await register.fillName("Test");
    await register.fillEmail("test+short@example.com");
    await register.fillPassword("short");
    await register.submit();
    await expect(page.getByText(/password must be at least/i)).toBeVisible();
  });
});
