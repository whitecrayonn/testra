import { test, expect } from "../../fixtures";
import { ProjectPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Projects @projects", () => {
  test("user can create and select a project through UI", async ({ authPage, workspace }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
    });

    const projects = new ProjectPage(authPage);
    await projects.create("E2E Project", "E2E", "Created by automated tests");
  });

  test("project list loads existing project", async ({ authPage, project, workspace }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const projects = new ProjectPage(authPage);
    await projects.goto();
    await expect(authPage.getByText(project.name)).toBeVisible();
  });

  test("happy path: create project via API", async ({ api, workspace }) => {
    const project = await api.createProject({
      workspace_id: workspace.id,
      name: "API Project",
      key: "APIPROJ",
    });

    const list = await api.listProjects(workspace.id);
    expect(list.some((p) => p.id === project.id)).toBeTruthy();
  });

  test("negative: create project with missing fields returns error", async ({ api, workspace }) => {
    await expect(
      api.createProject({ workspace_id: workspace.id, name: "", key: "" } as any),
    ).rejects.toThrow();
  });
});
