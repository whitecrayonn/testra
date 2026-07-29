import { test, expect } from "../../fixtures";
import { TestRunPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Manual Test Runner @manual @runs", () => {
  test("user can create a manual test run", async ({ authPage, workspace, project, testCase }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const runs = new TestRunPage(authPage);
    await runs.create(`Manual Run ${Date.now()}`, [testCase.id]);
  });

  test("manual run via API has pending status", async ({ api, workspace, project, testCase }) => {
    const run = await api.createTestRun({
      workspace_id: workspace.id,
      project_id: project.id,
      name: "API Manual Run",
      test_case_ids: [testCase.id],
    });

    const fetched = await api.getTestRun(run.id);
    expect(fetched.status).toMatch(/pending|running/);
    expect(fetched.total).toBeGreaterThanOrEqual(1);
  });

  test("negative: create run without test cases returns error", async ({ api, workspace, project }) => {
    await expect(
      api.createTestRun({ workspace_id: workspace.id, project_id: project.id, name: "Empty Run", test_case_ids: [] }),
    ).rejects.toThrow();
  });
});
