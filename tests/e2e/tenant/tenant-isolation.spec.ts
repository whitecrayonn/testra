import { test, expect, Browser } from "@playwright/test";
import { ApiHelper } from "../../helpers/api";
import { UserFactory, WorkspaceFactory, ProjectFactory, TestCaseFactory } from "../../factories";

test.describe("Tenant Isolation @tenant", () => {
  test("user cannot access workspace they do not belong to", async ({ browser, request }) => {
    const ownerApi = new ApiHelper(request);
    const ownerUser = await new UserFactory(ownerApi).create();
    const workspace = await new WorkspaceFactory(ownerApi).create();
    await ownerApi.login({ email: ownerUser.user.email, password: ownerUser.password });

    const otherContext = await (browser as Browser).newContext();
    const otherRequest = otherContext.request;
    const otherApi = new ApiHelper(otherRequest);
    const otherUser = await new UserFactory(otherApi).create();
    await otherApi.login({ email: otherUser.user.email, password: otherUser.password });

    await expect(otherApi.listProjects(workspace.id)).rejects.toThrow(/forbidden|not found|unauthorized/i);

    await otherContext.close();
  });

  test("resources are scoped to workspace", async ({ request }) => {
    const api = new ApiHelper(request);
    const user = await new UserFactory(api).create();
    await api.login({ email: user.user.email, password: user.password });

    const workspace1 = await new WorkspaceFactory(api).create();
    const project1 = await new ProjectFactory(api).create({ workspaceId: workspace1.id });
    const case1 = await new TestCaseFactory(api).create({ workspaceId: workspace1.id, projectId: project1.id });

    const workspace2 = await new WorkspaceFactory(api).create();
    const project2 = await new ProjectFactory(api).create({ workspaceId: workspace2.id });
    const case2 = await new TestCaseFactory(api).create({ workspaceId: workspace2.id, projectId: project2.id });

    const casesInFirst = await api.listTestCases(project1.id);
    expect(casesInFirst.data.some((c) => c.id === case2.id)).toBeFalsy();
    expect(casesInFirst.data.some((c) => c.id === case1.id)).toBeTruthy();

    const casesInSecond = await api.listTestCases(project2.id);
    expect(casesInSecond.data.some((c) => c.id === case1.id)).toBeFalsy();
    expect(casesInSecond.data.some((c) => c.id === case2.id)).toBeTruthy();
  });
});
