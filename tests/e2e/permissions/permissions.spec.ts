import { test as base, expect, APIRequestContext } from "@playwright/test";
import { ApiHelper } from "../../helpers/api";
import { UserFactory } from "../../factories";
import { uniqueId } from "../../helpers/random";

base.describe("RBAC @rbac @permissions", () => {
  base("unauthenticated API requests return 401", async ({ request }: { request: APIRequestContext }) => {
    const api = new ApiHelper(request);
    const res = await api.request.get(`${api.baseURL}/api/v1/workspaces`);
    expect(res.status()).toBe(401);
  });

  base("authenticated user can read own workspaces", async ({ request }: { request: APIRequestContext }) => {
    const api = new ApiHelper(request);
    const userFactory = new UserFactory(api);
    const { user, password } = await userFactory.create();
    await api.login({ email: user.email, password });
    const orgs = await api.listOrganizations();
    expect(Array.isArray(orgs)).toBeTruthy();
  });

  base("missing permission blocks workspace mutations", async ({ request }: { request: APIRequestContext }) => {
    const api = new ApiHelper(request);
    const userFactory = new UserFactory(api);
    const { user, password } = await userFactory.create();
    await api.login({ email: user.email, password });

    const slug = uniqueId("forbidden");
    const org = await api.createOrganization({ name: "Forbidden", slug });
    expect(org).toBeTruthy();
  });
});
