import { test, expect } from "../../fixtures";

test.describe("Integrations - Expanded @integrations @regression", () => {
  test("create integration via API", async ({ api, workspace }) => {
    const integration = await api.createIntegration({
      workspace_id: workspace.id,
      name: "Slack Integration",
      provider: "slack",
      config: { webhook_url: "https://hooks.slack.com/services/FAKE" },
    });
    expect(integration).toBeDefined();
  });

  test("list integrations via API", async ({ api, workspace }) => {
    await api.createIntegration({
      workspace_id: workspace.id,
      name: "List Test Integration",
      provider: "github",
      config: { webhook_url: "https://github.com/testra" },
    });

    const list = await api.listIntegrations(workspace.id);
    expect(list.data).toBeDefined();
    expect(Array.isArray(list.data)).toBeTruthy();
  });

  test("integration endpoints require workspace context", async ({ api }) => {
    await expect(api.listIntegrations("")).rejects.toThrow();
  });

  test("create integration without workspace fails", async ({ api }) => {
    await expect(
      api.createIntegration({
        workspace_id: "",
        name: "Should Fail",
        provider: "slack",
      }),
    ).rejects.toThrow();
  });

  test("create integration with empty name fails", async ({ api, workspace }) => {
    await expect(
      api.createIntegration({
        workspace_id: workspace.id,
        name: "",
        provider: "slack",
      }),
    ).rejects.toThrow();
  });

  test("create integration with empty provider fails", async ({ api, workspace }) => {
    await expect(
      api.createIntegration({
        workspace_id: workspace.id,
        name: "No Provider",
        provider: "",
      }),
    ).rejects.toThrow();
  });

  test("delete integration via API", async ({ api, workspace }) => {
    const integration = await api.createIntegration({
      workspace_id: workspace.id,
      name: "Delete Me Integration",
      provider: "slack",
    }) as any;

    if (integration.id) {
      await api.deleteIntegration(integration.id);
      await expect(api.getIntegration(integration.id)).rejects.toThrow();
    }
  });

  test("integrations are scoped to workspace", async ({ request }) => {
    const { ApiHelper } = await import("../../helpers/api");
    const { UserFactory, WorkspaceFactory } = await import("../../factories");

    const api1 = new ApiHelper(request);
    const user1 = await new UserFactory(api1).create();
    await api1.login({ email: user1.user.email, password: user1.password });
    const ws1 = await new WorkspaceFactory(api1).create();
    await api1.createIntegration({
      workspace_id: ws1.id,
      name: "Private Integration",
      provider: "slack",
    });

    const api2 = new ApiHelper(request);
    const user2 = await new UserFactory(api2).create();
    await api2.login({ email: user2.user.email, password: user2.password });
    const ws2 = await new WorkspaceFactory(api2).create();

    const list = await api2.listIntegrations(ws2.id);
    expect(list.data.some((i: any) => i.name === "Private Integration")).toBe(false);
  });

  test("create integration with different providers", async ({ api, workspace }) => {
    const providers = ["slack", "github", "jira", "email"];
    for (const provider of providers) {
      const integration = await api.createIntegration({
        workspace_id: workspace.id,
        name: `${provider} Integration`,
        provider,
      });
      expect(integration).toBeDefined();
    }
  });
});
