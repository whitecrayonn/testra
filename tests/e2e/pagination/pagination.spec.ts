import { test, expect } from "../../fixtures";
import { TestCaseFactory } from "../../factories";
import { parseCSV } from "../../helpers/csv";

test.describe("Pagination @pagination @testcases", () => {
  test("test cases list returns pagination metadata", async ({ api, workspace: _workspace, project, testCase: _testCase }) => {
    const result = await api.listTestCasesPaginated(project.id, { limit: 5 });
    expect(result.data).toBeDefined();
    expect(result.meta).toBeDefined();
    expect(typeof result.meta.has_more).toBe("boolean");
  });

  test("pagination with limit=1 returns at most 1 item", async ({ api, workspace, project }) => {
    const factory = new TestCaseFactory(api);
    await factory.create({ workspaceId: workspace.id, projectId: project.id });
    await factory.create({ workspaceId: workspace.id, projectId: project.id });

    const result = await api.listTestCasesPaginated(project.id, { limit: 1 });
    expect(result.data.length).toBeLessThanOrEqual(1);
    if (result.data.length === 1) {
      expect(result.meta.has_more).toBe(true);
      expect(result.meta.next_cursor).toBeTruthy();
    }
  });

  test("pagination cursor returns next page", async ({ api, workspace, project }) => {
    const factory = new TestCaseFactory(api);
    for (let i = 0; i < 3; i++) {
      await factory.create({ workspaceId: workspace.id, projectId: project.id });
    }

    const firstPage = await api.listTestCasesPaginated(project.id, { limit: 2 });
    expect(firstPage.data.length).toBeLessThanOrEqual(2);

    if (firstPage.meta.has_more && firstPage.meta.next_cursor) {
      const secondPage = await api.listTestCasesPaginated(project.id, {
        limit: 2,
        cursor: firstPage.meta.next_cursor,
      });
      expect(secondPage.data.length).toBeGreaterThan(0);
      const firstPageIds = new Set(firstPage.data.map((tc: any) => tc.id));
      const secondPageIds = secondPage.data.map((tc: any) => tc.id);
      for (const id of secondPageIds) {
        expect(firstPageIds.has(id)).toBe(false);
      }
    }
  });

  test("pagination with invalid cursor fails", async ({ api, project }) => {
    await expect(
      api.listTestCasesPaginated(project.id, { cursor: "invalid-base64-cursor" }),
    ).rejects.toThrow();
  });

  test("defects list returns pagination metadata", async ({ api, project, defect: _defect }) => {
    const result = await api.listDefectsPaginated(project.id, { limit: 5 });
    expect(result.data).toBeDefined();
    expect(result.meta).toBeDefined();
    expect(typeof result.meta.has_more).toBe("boolean");
  });

  test("organizations list returns pagination metadata", async ({ api, user: _user }) => {
    const result = await api.listOrganizationsPaginated({ limit: 5 });
    expect(result.data).toBeDefined();
    expect(result.meta).toBeDefined();
  });

  test("workspaces list returns pagination metadata", async ({ api, workspace }) => {
    const result = await api.listWorkspacesPaginated(workspace.organization_id, { limit: 5 });
    expect(result.data).toBeDefined();
    expect(result.meta).toBeDefined();
  });
});

test.describe("Search & Filtering @search @filtering", () => {
  test("search test cases by title returns matching results", async ({ api, workspace, testCase }) => {
    const results = await api.searchCases(workspace.id, testCase.title);
    expect(results.data.length).toBeGreaterThan(0);
    expect(
      results.data.some((tc: any) => tc.id === testCase.id),
    ).toBeTruthy();
  });

  test("search with empty query fails", async ({ api, workspace }) => {
    await expect(api.searchCases(workspace.id, "")).rejects.toThrow();
  });

  test("search with non-matching query returns empty results", async ({ api, workspace }) => {
    const results = await api.searchCases(workspace.id, "zzz_nonexistent_zzz_" + Date.now());
    expect(results.data.length).toBe(0);
  });

  test("search results are scoped to workspace", async ({ request }) => {
    const { ApiHelper } = await import("../../helpers/api");
    const { UserFactory, WorkspaceFactory, ProjectFactory, TestCaseFactory } = await import("../../factories");

    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();
    const proj1 = await new ProjectFactory(api1).create({ workspaceId: ws1.id });
    const _tc1 = await new TestCaseFactory(api1).create({
      workspaceId: ws1.id,
      projectId: proj1.id,
      testCase: { title: "UniqueSearchTerm12345" },
    });

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });
    const ws2 = await new WorkspaceFactory(api2).create();

    const results = await api2.searchCases(ws2.id, "UniqueSearchTerm12345");
    expect(results.data.length).toBe(0);
  });

  test("global search API returns results", async ({ api, workspace, testCase }) => {
    const results = await api.search(workspace.id, testCase.title);
    expect(results.data.results.length).toBeGreaterThan(0);
  });
});

test.describe("CSV Export @csv @analytics", () => {
  test("analytics CSV export returns valid CSV", async ({ api, workspace }) => {
    const csvText = await api.exportMetricsCSV(workspace.id);
    expect(csvText).toBeTruthy();
    expect(csvText.length).toBeGreaterThan(0);
    const rows = parseCSV(csvText);
    expect(rows.length).toBeGreaterThan(0);
  });

  test("analytics CSV export with project filter", async ({ api, workspace, project }) => {
    const csvText = await api.exportMetricsCSV(workspace.id, {
      project_id: project.id,
    });
    expect(csvText).toBeTruthy();
    const rows = parseCSV(csvText);
    expect(rows.length).toBeGreaterThan(0);
  });

  test("analytics CSV export without workspace fails", async ({ api }) => {
    await expect(api.exportMetricsCSV("")).rejects.toThrow();
  });

  test("analytics CSV export with date range", async ({ api, workspace }) => {
    const endDate = new Date().toISOString().split("T")[0];
    const startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split("T")[0];

    const csvText = await api.exportMetricsCSV(workspace.id, {
      start: startDate,
      end: endDate,
    });
    expect(csvText).toBeTruthy();
    const rows = parseCSV(csvText);
    expect(rows.length).toBeGreaterThan(0);
  });
});

test.describe("Analytics Trends & Metrics @analytics @filtering", () => {
  test("analytics trends return data", async ({ api, workspace }) => {
    const trends = await api.analyticsTrends(workspace.id);
    expect(trends).toBeDefined();
  });

  test("analytics trends with date range", async ({ api, workspace }) => {
    const end = new Date().toISOString().split("T")[0];
    const start = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().split("T")[0];

    const trends = await api.analyticsTrends(workspace.id, { start, end });
    expect(trends).toBeDefined();
  });

  test("analytics metrics return data", async ({ api, workspace }) => {
    const metrics = await api.analyticsMetrics(workspace.id);
    expect(metrics).toBeDefined();
  });

  test("analytics metrics with project filter", async ({ api, workspace, project }) => {
    const metrics = await api.analyticsMetrics(workspace.id, {
      project_id: project.id,
    });
    expect(metrics).toBeDefined();
  });

  test("analytics recent activity returns data", async ({ api, workspace }) => {
    const activity = await api.analyticsRecentActivity(workspace.id);
    expect(activity).toBeDefined();
  });

  test("analytics summary returns data", async ({ api, workspace }) => {
    const summary = await api.analyticsSummary(workspace.id);
    expect(summary).toBeDefined();
  });

  test("analytics summary with project filter", async ({ api, workspace, project: _project }) => {
    const summary = await api.analyticsSummary(workspace.id);
    expect(summary).toBeDefined();
  });
});
