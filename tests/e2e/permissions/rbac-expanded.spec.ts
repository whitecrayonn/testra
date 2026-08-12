import { test, expect } from "@playwright/test";
import { ApiHelper } from "../../helpers/api";
import { UserFactory, WorkspaceFactory, ProjectFactory, TestCaseFactory } from "../../factories";

test.describe("RBAC - Role-Based Access Control @rbac @permissions", () => {
  test("unauthenticated API request returns 401", async ({ request }) => {
    const api = new ApiHelper(request);
    await expect(api.me()).rejects.toThrow();
  });

  test("authenticated user can list their organizations", async ({ request }) => {
    const api = new ApiHelper(request);
    const user = await new UserFactory(api).create();
    await api.login({ email: user.user.email, password: user.password });
    const orgs = await api.listOrganizations();
    expect(Array.isArray(orgs)).toBeTruthy();
  });

  test("user cannot access another user's workspace resources", async ({ request }) => {
    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();
    const proj1 = await new ProjectFactory(api1).create({ workspaceId: ws1.id });
    const tc1 = await new TestCaseFactory(api1).create({
      workspaceId: ws1.id,
      projectId: proj1.id,
    });

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });

    await expect(api2.getTestCase(tc1.id)).rejects.toThrow();
  });

  test("user cannot create project in another user's workspace", async ({ request }) => {
    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });

    await expect(
      api2.createProject({
        workspace_id: ws1.id,
        name: "Cross-tenant Project",
        key: "CROSS",
      }),
    ).rejects.toThrow();
  });

  test("user cannot delete another user's project", async ({ request }) => {
    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();
    const proj1 = await new ProjectFactory(api1).create({ workspaceId: ws1.id });

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });

    await expect(api2.deleteProject(proj1.id)).rejects.toThrow();
  });

  test("user cannot update another user's test case", async ({ request }) => {
    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();
    const proj1 = await new ProjectFactory(api1).create({ workspaceId: ws1.id });
    const tc1 = await new TestCaseFactory(api1).create({
      workspaceId: ws1.id,
      projectId: proj1.id,
    });

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });

    await expect(
      api2.updateTestCase(tc1.id, { title: "Hacked Title" } as any),
    ).rejects.toThrow();
  });

  test("user cannot access another user's defects", async ({ request }) => {
    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();
    const proj1 = await new ProjectFactory(api1).create({ workspaceId: ws1.id });
    const defect1 = await api1.createDefect({
      workspace_id: ws1.id,
      project_id: proj1.id,
      title: "Private Defect",
    });

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });

    await expect(api2.getDefect(defect1.id)).rejects.toThrow();
  });

  test("workspace owner can access their workspace", async ({ request }) => {
    const api = new ApiHelper(request);
    const user = await new UserFactory(api).create();
    await api.login({ email: user.user.email, password: user.password });
    const ws = await new WorkspaceFactory(api).create();

    const fetched = await api.getWorkspace(ws.id);
    expect(fetched.id).toBe(ws.id);
  });

  test("integration endpoints require workspace context", async ({ request }) => {
    const api = new ApiHelper(request);
    const user = await new UserFactory(api).create();
    await api.login({ email: user.user.email, password: user.password });

    await expect(api.listIntegrations("")).rejects.toThrow();
  });

  test("notification creation requires workspace context", async ({ request }) => {
    const api = new ApiHelper(request);
    const user = await new UserFactory(api).create();
    await api.login({ email: user.user.email, password: user.password });

    await expect(
      api.createNotification({
        workspace_id: "",
        title: "Test Notification",
        message: "Test message",
      }),
    ).rejects.toThrow();
  });
});

test.describe("Tenant Isolation - Cross-Workspace @tenant @isolation", () => {
  test("workspaces are isolated between users", async ({ request }) => {
    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });

    await expect(api2.getWorkspace(ws1.id)).rejects.toThrow();
  });

  test("projects are isolated between workspaces", async ({ request }) => {
    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();
    const proj1 = await new ProjectFactory(api1).create({ workspaceId: ws1.id });

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });
    const ws2 = await new WorkspaceFactory(api2).create();

    const projects = await api2.listProjects(ws2.id);
    expect(projects.some((p: any) => p.id === proj1.id)).toBe(false);
  });

  test("test cases are scoped to their workspace", async ({ request }) => {
    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();
    const proj1 = await new ProjectFactory(api1).create({ workspaceId: ws1.id });
    const tc1 = await new TestCaseFactory(api1).create({
      workspaceId: ws1.id,
      projectId: proj1.id,
    });

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });
    const ws2 = await new WorkspaceFactory(api2).create();
    const proj2 = await new ProjectFactory(api2).create({ workspaceId: ws2.id });

    const testCases = await api2.listTestCases(proj2.id);
    expect(testCases.data.some((tc: any) => tc.id === tc1.id)).toBe(false);
  });

  test("automation projects are scoped to workspace", async ({ request }) => {
    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();
    const auto1 = await api1.createAutomationProject({
      workspace_id: ws1.id,
      name: "Isolated Automation",
      framework: "playwright",
    });

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });
    const ws2 = await new WorkspaceFactory(api2).create();

    const list = await api2.listAutomationProjects(ws2.id);
    expect(list.data.some((a: any) => a.id === auto1.id)).toBe(false);
  });

  test("api collections are scoped to workspace", async ({ request }) => {
    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();
    const collection1 = await api1.createApiCollection({
      workspace_id: ws1.id,
      name: "Isolated Collection",
    });

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });

    await expect(api2.getApiCollection(collection1.id)).rejects.toThrow();
  });

  test("search results are scoped to workspace", async ({ request }) => {
    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();
    const proj1 = await new ProjectFactory(api1).create({ workspaceId: ws1.id });
    const _tc1 = await new TestCaseFactory(api1).create({
      workspaceId: ws1.id,
      projectId: proj1.id,
      testCase: { title: "TenantIsolationUniqueTerm789" },
    });

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });
    const ws2 = await new WorkspaceFactory(api2).create();

    const results = await api2.searchCases(ws2.id, "TenantIsolationUniqueTerm789");
    expect(results.data.length).toBe(0);
  });
});
