import { test, expect } from "../../fixtures";
import { TestCasePage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Test Cases @testcases", () => {
  test("user can create a test case through UI", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const testCases = new TestCasePage(authPage);
    await testCases.create("Verify login flow", "Ensure user can log in", ["Open login", "Enter credentials"]);
  });

  test("test case search returns expected result", async ({ authPage, workspace, project, testCase }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const testCases = new TestCasePage(authPage);
    await testCases.search(testCase.title);
    await testCases.expectTestCase(testCase.title);
  });

  test("CRUD via API", async ({ api, workspace, project }) => {
    const created = await api.createTestCase({
      workspace_id: workspace.id,
      project_id: project.id,
      title: "API Test Case",
      priority: "high",
      tags: ["api"],
      steps: [{ action: "Call API", expected: "200 OK" }],
    });

    expect(created.title).toBe("API Test Case");

    const fetched = await api.getTestCase(created.id);
    expect(fetched.id).toBe(created.id);

    await api.updateTestCase(created.id, { title: "Updated API Test Case" });
    const updated = await api.getTestCase(created.id);
    expect(updated.title).toBe("Updated API Test Case");

    await api.deleteTestCase(created.id);
    await expect(api.getTestCase(created.id)).rejects.toThrow();
  });

  test("negative: create test case without project fails", async ({ api, workspace }) => {
    await expect(
      api.createTestCase({ workspace_id: workspace.id, project_id: "", title: "Orphan" } as any),
    ).rejects.toThrow();
  });
});
