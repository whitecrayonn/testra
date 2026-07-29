import { test, expect } from "@playwright/test";
import { ApiHelper } from "../../helpers/api";
import { UserFactory, WorkspaceFactory, ProjectFactory, AutomationProjectFactory } from "../../factories";
import { setWorkspaceContext } from "../../helpers/storage";
import path from "path";

test.describe("File Upload @upload @automation", () => {
  test("upload JUnit XML report to automation project", async ({ browser, request }) => {
    const api = new ApiHelper(request);
    const user = await new UserFactory(api).create();
    await api.login({ email: user.user.email, password: user.password });
    const workspace = await new WorkspaceFactory(api).create();
    const project = await new ProjectFactory(api).create({ workspaceId: workspace.id });
    const autoProject = await new AutomationProjectFactory(api).create({
      workspaceId: workspace.id,
      projectId: project.id,
      automationProject: { name: "Upload Target", framework: "junit" },
    });

    const context = await browser.newContext();
    await context.request.storageState({ path: undefined });

    await api.request.storageState({ path: path.join(test.info().outputDir, "upload-state.json") });
    const authContext = await browser.newContext({
      storageState: path.join(test.info().outputDir, "upload-state.json"),
    });
    const authPage = await authContext.newPage();

    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    await authPage.goto(`/dashboard/automation/${autoProject.id}`);
    await expect(authPage.getByText(autoProject.name)).toBeVisible();
    await authPage.getByRole("button", { name: "Import Report" }).click();
    await authPage.getByPlaceholder("Report name").fill("JUnit Upload");
    await authPage.locator('input[type="file"]').setInputFiles(path.join(__dirname, "../../test-data/sample-junit.xml"));
    await authPage.getByRole("button", { name: "Import", exact: true }).click();

    await expect(authPage.getByText(/imported|execution|success/i).first()).toBeVisible();

    await authContext.close();
    await context.close();
  });
});
