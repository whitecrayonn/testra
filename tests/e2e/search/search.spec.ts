import { test, expect } from "../../fixtures";
import { SearchPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Global Search @search", () => {
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
});
