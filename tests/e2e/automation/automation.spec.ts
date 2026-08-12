import { test, expect } from "../../fixtures";
import { AutomationPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Automation Hub @automation", () => {
  test("user can create an automation project through UI", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const automation = new AutomationPage(authPage);
    await automation.create(`Automation ${Date.now()}`, "playwright");
  });

  test("automation project via API stores framework", async ({ api, workspace, project }) => {
    const projectAuto = await api.createAutomationProject({
      workspace_id: workspace.id,
      project_id: project.id,
      name: "API Automation",
      framework: "junit",
      repository_url: "https://github.com/example/repo",
      branch: "main",
      command: "npx playwright test",
    });

    expect(projectAuto.framework).toBe("junit");

    const list = await api.listAutomationProjects(workspace.id);
    expect(list.data.some((p) => p.id === projectAuto.id)).toBeTruthy();
  });

  test("negative: create automation project with unsupported framework", async ({ api, workspace }) => {
    await expect(
      api.createAutomationProject({
        workspace_id: workspace.id,
        name: "Bad Framework",
        framework: "unknown",
      } as any),
    ).rejects.toThrow();
  });
});
