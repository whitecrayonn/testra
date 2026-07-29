import { test, expect } from "../../fixtures";

test.describe("Integration Hub @integrations", () => {
  test("create and list integration via API", async ({ api, workspace }) => {
    const integration = await api.createIntegration({
      workspace_id: workspace.id,
      name: "Slack Integration",
      provider: "slack",
      config: { webhook_url: "https://hooks.slack.com/services/FAKE" },
    });

    expect(integration).toBeDefined();

    const list = await api.listIntegrations(workspace.id);
    expect(list.data.some((i: any) => (i as any).id === (integration as any).id || (i as any).name === "Slack Integration")).toBeTruthy();
  });

  test("integration endpoints require workspace context", async ({ api }) => {
    await expect(api.listIntegrations("")).rejects.toThrow();
  });
});
