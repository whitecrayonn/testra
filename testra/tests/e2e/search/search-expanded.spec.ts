import { test, expect } from "../../fixtures";
import { SearchPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";
import { TestCaseFactory } from "../../factories";

test.describe("Search - Expanded @search @regression", () => {
  test("search returns test case via API", async ({ api, workspace, testCase }) => {
    const results = await api.search(workspace.id, testCase.title);
    expect(results.data.results.length).toBeGreaterThan(0);
  });

  test("search UI displays results", async ({ authPage, workspace, project, testCase }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const search = new SearchPage(authPage);
    await authPage.goto("/dashboard");
    await search.search(testCase.title);
  });

  test("search via test-cases search endpoint returns matching cases", async ({ api, workspace, testCase }) => {
    const results = await api.searchCases(workspace.id, testCase.title);
    expect(results.data.length).toBeGreaterThan(0);
    expect(results.data.some((tc: any) => tc.id === testCase.id)).toBeTruthy();
  });

  test("search with partial title matches", async ({ api, workspace, project }) => {
    const factory = new TestCaseFactory(api);
    const tc = await factory.create({
      workspaceId: workspace.id,
      projectId: project.id,
      testCase: { title: "PartialMatchTestUnique123" },
    });

    const results = await api.searchCases(workspace.id, "PartialMatch");
    expect(results.data.length).toBeGreaterThan(0);
    expect(results.data.some((r: any) => r.id === tc.id)).toBeTruthy();
  });

  test("search is case-insensitive", async ({ api, workspace, project }) => {
    const factory = new TestCaseFactory(api);
    const tc = await factory.create({
      workspaceId: workspace.id,
      projectId: project.id,
      testCase: { title: "CaseInsensitiveSearchTest" },
    });

    const results = await api.searchCases(workspace.id, "caseinsensitivesearchtest");
    expect(results.data.some((r: any) => r.id === tc.id)).toBeTruthy();
  });

  test("search with special characters does not crash", async ({ api, workspace }) => {
    await expect(
      api.searchCases(workspace.id, "test!@#$%^&*()"),
    ).resolves.toBeDefined();
  });

  test("search with very long query does not crash", async ({ api, workspace }) => {
    const longQuery = "a".repeat(500);
    await expect(
      api.searchCases(workspace.id, longQuery),
    ).resolves.toBeDefined();
  });

  test("search results include pagination metadata", async ({ api, workspace, project }) => {
    const factory = new TestCaseFactory(api);
    await factory.create({
      workspaceId: workspace.id,
      projectId: project.id,
      testCase: { title: "PaginationSearchTest" },
    });

    const results = await api.searchCases(workspace.id, "PaginationSearchTest", { limit: 1 });
    expect(results.meta).toBeDefined();
    expect(typeof results.meta.has_more).toBe("boolean");
  });

  test("global search API returns results for test cases", async ({ api, workspace, testCase }) => {
    const results = await api.search(workspace.id, testCase.title);
    expect(results.data.results.length).toBeGreaterThan(0);
  });

  test("global search API returns empty for non-matching query", async ({ api, workspace }) => {
    const results = await api.search(workspace.id, "zzz_nonexistent_zzz_" + Date.now());
    expect(results.data.results.length).toBe(0);
  });

  test("search without workspace_id fails", async ({ api }) => {
    await expect(api.search("", "test")).rejects.toThrow();
  });
});
