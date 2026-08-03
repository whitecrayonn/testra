import { test, expect } from "../../fixtures";
import { AnalyticsPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Analytics @analytics", () => {
  test("analytics summary API returns data", async ({ api, workspace, project: _project, testCase: _testCase, testRun: _testRun }) => {
    const summary = await api.analyticsSummary(workspace.id);
    expect(summary).toBeDefined();
  });

  test("executive dashboard loads", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const analytics = new AnalyticsPage(authPage);
    await analytics.goto();
    await analytics.expectSummaryVisible();
  });
});
