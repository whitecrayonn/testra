import { test, expect } from "../../fixtures";
import { DefectPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Defects @defects", () => {
  test("user can create a defect through UI", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const defects = new DefectPage(authPage);
    await defects.create("Login button unresponsive", "Clicking login does nothing", "high", "critical");
  });

  test("defect via API carries severity and priority", async ({ api, workspace, project }) => {
    const defect = await api.createDefect({
      workspace_id: workspace.id,
      project_id: project.id,
      title: "API Defect",
      severity: "critical",
      priority: "high",
    });

    expect(defect.severity).toBe("critical");
    expect(defect.priority).toBe("high");

    const list = await api.listDefects(project.id);
    expect(list.data.some((d) => d.id === defect.id)).toBeTruthy();
  });

  test("negative: create defect without project fails", async ({ api, workspace }) => {
    await expect(
      api.createDefect({ workspace_id: workspace.id, project_id: "", title: "Orphan" } as any),
    ).rejects.toThrow();
  });
});
