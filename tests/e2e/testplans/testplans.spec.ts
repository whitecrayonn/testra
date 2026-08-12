import { test, expect } from "../../fixtures";
import { TestPlanPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Test Plans @testplans", () => {
  test("user can create a test plan through UI", async ({ authPage, workspace, project, testCase }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const plans = new TestPlanPage(authPage);
    await plans.create(`Plan ${Date.now()}`, [testCase.id]);
  });

  test("test plan via API includes test cases", async ({ api, workspace, project, testCase }) => {
    const plan = await api.createTestPlan({
      workspace_id: workspace.id,
      project_id: project.id,
      name: "API Plan",
      test_case_ids: [testCase.id],
    });

    const fetched = await api.getTestPlan(plan.id);
    expect(fetched.test_case_ids).toContain(testCase.id);
  });

  test("test plans list and detail pages render without crashing", async ({
    authPage,
    api,
    workspace,
    project,
    testCase,
  }) => {
    const plan = await api.createTestPlan({
      workspace_id: workspace.id,
      project_id: project.id,
      name: `Regression Plan ${Date.now()}`,
      test_case_ids: [testCase.id],
    });

    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    await authPage.goto("/dashboard/test-plans");
    await expect(authPage.getByText(plan.name)).toBeVisible();
    await expect(authPage.getByText(/something went wrong/i)).not.toBeVisible();

    await authPage.goto(`/dashboard/test-plans/${plan.id}`);
    await expect(authPage.getByRole("heading", { name: /test cases/i })).toBeVisible();
    await expect(authPage.getByText(/something went wrong/i)).not.toBeVisible();
  });

  test("negative: create test plan with invalid test case ids", async ({ api, workspace, project }) => {
    await expect(
      api.createTestPlan({
        workspace_id: workspace.id,
        project_id: project.id,
        name: "Bad Plan",
        test_case_ids: ["00000000-0000-0000-0000-000000000000"],
      }),
    ).rejects.toThrow();
  });
});
