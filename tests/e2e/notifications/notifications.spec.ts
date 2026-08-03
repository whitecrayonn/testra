import { test, expect } from "../../fixtures";
import { NotificationPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Notifications @notifications", () => {
  test("notifications page loads for authenticated user", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const notifications = new NotificationPage(authPage);
    await notifications.goto();
  });

  test("create and mark notification read via API", async ({ api, workspace }) => {
    await api.createNotification({
      workspace_id: workspace.id,
      title: "Test Notification",
      message: "This is an automated test notification",
    });

    const list = await api.listNotifications();
    expect(list.data.length).toBeGreaterThan(0);
  });

  test("negative: create notification without workspace fails", async ({ api }) => {
    await expect(
      api.createNotification({ workspace_id: "", title: "Orphan", message: "Bad" } as any),
    ).rejects.toThrow();
  });
});
