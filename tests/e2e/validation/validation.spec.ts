import { test, expect } from "../../fixtures";
import { ApiHelper } from "../../helpers/api";
import { UserFactory } from "../../factories";

test.describe("Input Validation @validation @negative", () => {
  test("register with invalid email fails", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(
      api.register({ email: "not-an-email", password: "SecurePass123!", name: "Test User" }),
    ).rejects.toThrow();
  });

  test("register with empty email fails", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(
      api.register({ email: "", password: "SecurePass123!", name: "Test User" }),
    ).rejects.toThrow();
  });

  test("register with empty name fails", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(
      api.register({ email: "valid@example.com", password: "SecurePass123!", name: "" }),
    ).rejects.toThrow();
  });

  test("register with short password fails", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(
      api.register({ email: "valid@example.com", password: "123", name: "Test User" }),
    ).rejects.toThrow();
  });

  test("register with no uppercase password fails", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(
      api.register({ email: "valid@example.com", password: "lowercase123!", name: "Test User" }),
    ).rejects.toThrow();
  });

  test("register with no number password fails", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(
      api.register({ email: "valid@example.com", password: "NoNumberHere!", name: "Test User" }),
    ).rejects.toThrow();
  });

  test("login with non-existent user fails", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(
      api.login({ email: "nonexistent@example.com", password: "SecurePass123!" }),
    ).rejects.toThrow();
  });

  test("login with wrong password fails", async ({ request }) => {
    const api = new ApiHelper(request);
    const user = await new UserFactory(api).create();
    await expect(
      api.login({ email: user.user.email, password: "WrongPassword123!" }),
    ).rejects.toThrow();
  });

  test("login with empty credentials fails", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(
      api.login({ email: "", password: "" }),
    ).rejects.toThrow();
  });

  test("duplicate registration fails", async ({ request }) => {
    const api = new ApiHelper(request);
    const user = await new UserFactory(api).create();
    await expect(
      api.register({ email: user.user.email, password: "SecurePass123!", name: "Duplicate User" }),
    ).rejects.toThrow();
  });

  test("create organization with empty name fails", async ({ request }) => {
    const api = new ApiHelper(request);
    const user = await new UserFactory(api).create();
    await api.login({ email: user.user.email, password: user.password });
    await expect(
      api.createOrganization({ name: "", slug: "test-slug" }),
    ).rejects.toThrow();
  });

  test("create workspace with invalid organization ID fails", async ({ request }) => {
    const api = new ApiHelper(request);
    const user = await new UserFactory(api).create();
    await api.login({ email: user.user.email, password: user.password });
    await expect(
      api.createWorkspace({
        organization_id: "not-a-uuid",
        name: "Test WS",
        slug: "test-ws",
      }),
    ).rejects.toThrow();
  });

  test("create project with invalid workspace ID fails", async ({ request }) => {
    const api = new ApiHelper(request);
    const user = await new UserFactory(api).create();
    await api.login({ email: user.user.email, password: user.password });
    await expect(
      api.createProject({
        workspace_id: "not-a-uuid",
        name: "Test Project",
        key: "TEST",
      }),
    ).rejects.toThrow();
  });

  test("create test case with invalid UUID fails", async ({ request }) => {
    const api = new ApiHelper(request);
    const user = await new UserFactory(api).create();
    await api.login({ email: user.user.email, password: user.password });
    await expect(
      api.createTestCase({
        workspace_id: "not-a-uuid",
        project_id: "not-a-uuid",
        title: "Test Case",
      }),
    ).rejects.toThrow();
  });

  test("create defect with invalid severity fails", async ({ api, workspace, project }) => {
    await expect(
      api.createDefect({
        workspace_id: workspace.id,
        project_id: project.id,
        title: "Test Defect",
        severity: "invalid_severity",
      }),
    ).rejects.toThrow();
  });

  test("create automation project with empty name fails", async ({ api, workspace }) => {
    await expect(
      api.createAutomationProject({
        workspace_id: workspace.id,
        name: "",
        framework: "playwright",
      }),
    ).rejects.toThrow();
  });

  test("create api collection with empty name fails", async ({ api, workspace }) => {
    await expect(
      api.createApiCollection({
        workspace_id: workspace.id,
        name: "",
      }),
    ).rejects.toThrow();
  });

  test("create test plan with empty name fails", async ({ api, workspace, project }) => {
    await expect(
      api.createTestPlan({
        workspace_id: workspace.id,
        project_id: project.id,
        name: "",
      }),
    ).rejects.toThrow();
  });

  test("create test run with empty name fails", async ({ api, workspace, project }) => {
    await expect(
      api.createTestRun({
        workspace_id: workspace.id,
        project_id: project.id,
        name: "",
      }),
    ).rejects.toThrow();
  });

  test("analytics summary without workspace_id fails", async ({ api }) => {
    await expect(api.analyticsSummary("")).rejects.toThrow();
  });

  test("analytics trends without workspace_id fails", async ({ api }) => {
    await expect(api.analyticsTrends("")).rejects.toThrow();
  });

  test("analytics metrics without workspace_id fails", async ({ api }) => {
    await expect(api.analyticsMetrics("")).rejects.toThrow();
  });

  test("analytics with invalid date format fails", async ({ api, workspace }) => {
    await expect(
      api.analyticsTrends(workspace.id, { start: "invalid-date", end: "also-invalid" }),
    ).rejects.toThrow();
  });
});
