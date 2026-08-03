import { test, expect } from "../../fixtures";
import { NotificationPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("Notifications - Expanded @notifications @regression", () => {
  test("notifications page loads for authenticated user", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const notifPage = new NotificationPage(authPage);
    await notifPage.goto();
    await expect(authPage).toHaveURL(/.*notifications.*/);
  });

  test("create notification via API", async ({ api, workspace }) => {
    const notification = await api.createNotification({
      workspace_id: workspace.id,
      title: "Test Notification",
      message: "This is a test notification message",
    });
    expect(notification).toBeDefined();
  });

  test("list notifications via API", async ({ api, workspace }) => {
    await api.createNotification({
      workspace_id: workspace.id,
      title: "List Test Notification",
      message: "Testing list endpoint",
    });

    const list = await api.listNotifications();
    expect(list.data).toBeDefined();
    expect(Array.isArray(list.data)).toBeTruthy();
  });

  test("create notification without workspace fails", async ({ api }) => {
    await expect(
      api.createNotification({
        workspace_id: "",
        title: "Should Fail",
        message: "No workspace",
      }),
    ).rejects.toThrow();
  });

  test("create notification with empty title fails", async ({ api, workspace }) => {
    await expect(
      api.createNotification({
        workspace_id: workspace.id,
        title: "",
        message: "Empty title",
      }),
    ).rejects.toThrow();
  });

  test("create notification with empty message fails", async ({ api, workspace }) => {
    await expect(
      api.createNotification({
        workspace_id: workspace.id,
        title: "Empty Message Test",
        message: "",
      }),
    ).rejects.toThrow();
  });

  test("mark notification as read via API", async ({ api, workspace }) => {
    const notification = await api.createNotification({
      workspace_id: workspace.id,
      title: "Mark Read Test",
      message: "Testing mark as read",
    }) as any;

    if (notification.id) {
      await api.markNotificationRead(notification.id);
    }
  });

  test("notifications are scoped to workspace", async ({ request }) => {
    const { ApiHelper } = await import("../../helpers/api");
    const { UserFactory, WorkspaceFactory } = await import("../../factories");

    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();
    await api1.createNotification({
      workspace_id: ws1.id,
      title: "Private Notification",
      message: "Should not be visible to other users",
    });

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });
    await new WorkspaceFactory(api2).create();

    const list = await api2.listNotifications();
    expect(list.data.some((n: any) => n.title === "Private Notification")).toBe(false);
  });
});
