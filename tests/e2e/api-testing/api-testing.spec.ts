import { test, expect } from "../../fixtures";
import { ApiTestingPage } from "../../pages";
import { setWorkspaceContext } from "../../helpers/storage";

test.describe("API Testing Module @api @apitests", () => {
  test("user can create and view an API collection", async ({ authPage, workspace, project }) => {
    await setWorkspaceContext(authPage, {
      workspaceId: workspace.id,
      workspaceName: workspace.name,
      organizationId: workspace.organization_id,
      projectId: project.id,
      projectName: project.name,
    });

    const apiTesting = new ApiTestingPage(authPage);
    await apiTesting.createCollection(`Collection ${Date.now()}`);
  });

  test("API collection CRUD via API", async ({ api, workspace }) => {
    const collection = await api.createApiCollection({
      workspace_id: workspace.id,
      name: "API Collection",
    });

    const environment = await api.createApiEnvironment({
      workspace_id: workspace.id,
      name: "API Environment",
      variables: { baseUrl: "https://jsonplaceholder.typicode.com" },
    });

    const request = await api.createApiRequest({
      collection_id: collection.id,
      name: "Get Post",
      method: "GET",
      url: "https://jsonplaceholder.typicode.com/posts/1",
    });

    expect(request.collection_id).toBe(collection.id);

    const execution = await api.executeApiRequest({
      workspace_id: workspace.id,
      request_id: request.id,
      environment_id: environment.id,
    });

    expect(execution).toBeDefined();
  });

  test("negative: execute request with invalid URL", async ({ api, workspace }) => {
    const collection = await api.createApiCollection({ workspace_id: workspace.id, name: "Bad Collection" });
    const request = await api.createApiRequest({
      collection_id: collection.id,
      name: "Bad Request",
      method: "GET",
      url: "not-a-url",
    });

    await expect(
      api.executeApiRequest({ workspace_id: workspace.id, request_id: request.id }),
    ).rejects.toThrow();
  });
});
