import { test, expect } from "../../fixtures";

test.describe("API Integration @api @integration", () => {
  test("execute API request against external echo service", async ({ api, workspace }) => {
    const collection = await api.createApiCollection({ workspace_id: workspace.id, name: "Integration Collection" });
    const environment = await api.createApiEnvironment({
      workspace_id: workspace.id,
      name: "Integration Env",
      variables: { baseUrl: "https://httpbin.org" },
    });
    const request = await api.createApiRequest({
      collection_id: collection.id,
      name: "GET test",
      method: "GET",
      url: "https://httpbin.org/get",
    });

    const execution = await api.executeApiRequest({
      workspace_id: workspace.id,
      request_id: request.id,
      environment_id: environment.id,
    });

    expect(execution).toBeDefined();
  });

  test("collection and environment are scoped to workspace", async ({ api: _api, workspace, apiCollection }) => {
    expect(apiCollection.collection.workspace_id).toBe(workspace.id);
    expect(apiCollection.environment.workspace_id).toBe(workspace.id);
    expect(apiCollection.request.collection_id).toBe(apiCollection.collection.id);
  });
});
