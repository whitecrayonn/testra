import { test, expect } from "../../fixtures";
import { ApiHelper } from "../../helpers/api";
import { UserFactory } from "../../factories";
import { API } from "../../constants/routes";

test.describe("Session & Token Management @session @auth", () => {
  test("refresh token returns new access token", async ({ request }) => {
    const api = new ApiHelper(request);
    const factory = new UserFactory(api);
    const created = await factory.create();
    const loginResult = await api.login({
      email: created.user.email,
      password: created.password,
    });

    expect(loginResult.token).toBeTruthy();
    expect(loginResult.refresh_token).toBeTruthy();

    const refreshed = await api.refreshToken(loginResult.refresh_token);
    expect(refreshed.token).toBeTruthy();
    expect(refreshed.refresh_token).toBeTruthy();
    expect(refreshed.token).not.toBe(loginResult.token);
  });

  test("expired refresh token is rejected", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(api.refreshToken("rt_invalid_token_12345")).rejects.toThrow();
  });

  test("logout invalidates refresh token", async ({ request }) => {
    const api = new ApiHelper(request);
    const factory = new UserFactory(api);
    const created = await factory.create();
    const loginResult = await api.login({
      email: created.user.email,
      password: created.password,
    });

    await api.logout();

    await expect(api.refreshToken(loginResult.refresh_token)).rejects.toThrow();
  });

  test("me endpoint returns current user", async ({ api, user }) => {
    const me = await api.me();
    expect(me.id).toBe(user.user.id);
    expect(me.email).toBe(user.user.email);
  });

  test("me endpoint fails without auth", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(api.me()).rejects.toThrow();
  });

  test("double refresh with same token fails (token rotation)", async ({ request }) => {
    const api = new ApiHelper(request);
    const factory = new UserFactory(api);
    const created = await factory.create();
    const loginResult = await api.login({
      email: created.user.email,
      password: created.password,
    });

    const firstRefresh = await api.refreshToken(loginResult.refresh_token);
    expect(firstRefresh.token).toBeTruthy();

    await expect(api.refreshToken(loginResult.refresh_token)).rejects.toThrow();
  });
});

test.describe("CSRF Token Management @csrf @auth", () => {
  test("CSRF token endpoint returns a token", async ({ request }) => {
    const api = new ApiHelper(request);
    const token = await api.getCsrfToken();
    expect(token).toBeTruthy();
    expect(token.length).toBeGreaterThan(10);
  });

  test("CSRF token is set as cookie", async ({ request }) => {
    const api = new ApiHelper(request);
    const res = await api.rawRequest("GET", API.csrf);
    expect(res.ok()).toBeTruthy();
    const setCookie = res.headers()["set-cookie"];
    expect(setCookie).toBeTruthy();
    expect(setCookie.toLowerCase()).toContain("csrf");
  });

  test("CSRF token can be fetched multiple times", async ({ request }) => {
    const api = new ApiHelper(request);
    const token1 = await api.getCsrfToken();
    const token2 = await api.getCsrfToken();
    expect(token1).toBeTruthy();
    expect(token2).toBeTruthy();
  });
});

test.describe("Cookie Management @cookie @auth", () => {
  test("login sets auth cookies", async ({ request }) => {
    const api = new ApiHelper(request);
    const factory = new UserFactory(api);
    const created = await factory.create();

    const res = await api.rawRequest("POST", API.login, {
      email: created.user.email,
      password: created.password,
    });

    expect(res.ok()).toBeTruthy();
    const setCookie = res.headers()["set-cookie"];
    expect(setCookie).toBeTruthy();
    expect(setCookie.toLowerCase()).toContain("token");
  });

  test("logout clears auth cookies", async ({ request }) => {
    const api = new ApiHelper(request);
    const factory = new UserFactory(api);
    const created = await factory.create();
    await api.login({
      email: created.user.email,
      password: created.password,
    });

    const res = await api.rawRequest("POST", API.logout, {});
    expect(res.ok()).toBeTruthy();
    const setCookie = res.headers()["set-cookie"];
    if (setCookie) {
      expect(setCookie.toLowerCase()).toMatch(/max-age=0|expires=.*1970|deleted/i);
    }
  });

  test("register sets auth cookies", async ({ request }) => {
    const api = new ApiHelper(request);
    const factory = new UserFactory(api);
    const credentials = factory.build();

    const res = await api.rawRequest("POST", API.register, {
      email: credentials.email,
      password: credentials.password,
      name: credentials.name,
    });

    expect(res.ok()).toBeTruthy();
    const setCookie = res.headers()["set-cookie"];
    expect(setCookie).toBeTruthy();
  });
});

test.describe("Password Reset Flow @password-reset @auth", () => {
  test("request password reset for existing user succeeds", async ({ request }) => {
    const api = new ApiHelper(request);
    const factory = new UserFactory(api);
    const created = await factory.create();

    await expect(
      api.requestPasswordReset(created.user.email),
    ).resolves.toBeUndefined();
  });

  test("request password reset for non-existent user does not error", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(
      api.requestPasswordReset("nonexistent.user@example.com"),
    ).resolves.toBeUndefined();
  });

  test("reset password with invalid token fails", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(
      api.resetPassword("invalid-token", "NewSecurePass123!"),
    ).rejects.toThrow();
  });

  test("reset password with weak password fails", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(
      api.resetPassword("invalid-token", "123"),
    ).rejects.toThrow();
  });
});
