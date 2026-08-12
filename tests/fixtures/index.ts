import { test as base, expect, Page } from "@playwright/test";
import path from "path";
import { ApiHelper } from "../helpers/api";
import {
  UserFactory,
  WorkspaceFactory,
  ProjectFactory,
  TestCaseFactory,
  TestPlanFactory,
  TestRunFactory,
  DefectFactory,
  AutomationProjectFactory,
  ApiCollectionFactory,
} from "../factories";
import {
  User,
  Workspace,
  Project,
  TestCase,
  TestPlan,
  TestRun,
  Defect,
  AutomationProject,
  ApiCollection,
  ApiEnvironment,
  ApiRequest,
} from "../types";

export interface CreatedUser {
  user: User;
  password: string;
  storageState: string;
}

interface TestFixtures {
  api: ApiHelper;
  user: CreatedUser;
  authPage: Page;
  workspace: Workspace;
  project: Project;
  testCase: TestCase;
  testPlan: TestPlan;
  testRun: TestRun;
  defect: Defect;
  automationProject: AutomationProject;
  apiCollection: {
    collection: ApiCollection;
    environment: ApiEnvironment;
    request: ApiRequest;
  };
  authenticatedUser: CreatedUser;
  adminUser: CreatedUser;
  qaUser: CreatedUser;
  viewerUser: CreatedUser;
  workspaceOwner: CreatedUser;
}

async function createUserWithStorage(
  api: ApiHelper,
  testInfo: { outputDir: string },
  prefix: string,
): Promise<CreatedUser> {
  const factory = new UserFactory(api);
  const created = await factory.create();
  const storagePath = path.join(testInfo.outputDir, `${prefix}-auth-storage.json`);
  await api.request.storageState({ path: storagePath });
  return { user: created.user, password: created.password, storageState: storagePath };
}

export const test = base.extend<TestFixtures>({
  api: async ({ request }, use, _testInfo) => {
    const api = new ApiHelper(request, process.env.TEST_API_URL || "http://localhost:8080");
    await use(api);
    await request.dispose();
  },

  user: async ({ api }, use, testInfo) => {
    const created = await createUserWithStorage(api, testInfo, "user");
    await use(created);
  },

  authPage: async ({ browser, user }, use) => {
    const context = await browser.newContext({ storageState: user.storageState });
    const page = await context.newPage();
    await use(page);
    await context.close();
  },

  workspace: async ({ api, user: _user }, use) => {
    const factory = new WorkspaceFactory(api);
    const workspace = await factory.create();
    await use(workspace);
  },

  project: async ({ api, workspace }, use) => {
    const factory = new ProjectFactory(api);
    const project = await factory.create({ workspaceId: workspace.id });
    await use(project);
  },

  testCase: async ({ api, workspace, project }, use) => {
    const factory = new TestCaseFactory(api);
    const testCase = await factory.create({
      workspaceId: workspace.id,
      projectId: project.id,
    });
    await use(testCase);
  },

  testPlan: async ({ api, workspace, project, testCase }, use) => {
    const factory = new TestPlanFactory(api);
    const testPlan = await factory.create({
      workspaceId: workspace.id,
      projectId: project.id,
      testCaseIds: [testCase.id],
    });
    await use(testPlan);
  },

  testRun: async ({ api, workspace, project, testCase }, use) => {
    const factory = new TestRunFactory(api);
    const testRun = await factory.create({
      workspaceId: workspace.id,
      projectId: project.id,
      testCaseIds: [testCase.id],
    });
    await use(testRun);
  },

  defect: async ({ api, workspace, project }, use) => {
    const factory = new DefectFactory(api);
    const defect = await factory.create({
      workspaceId: workspace.id,
      projectId: project.id,
    });
    await use(defect);
  },

  automationProject: async ({ api, workspace, project }, use) => {
    const factory = new AutomationProjectFactory(api);
    const automationProject = await factory.create({
      workspaceId: workspace.id,
      projectId: project.id,
    });
    await use(automationProject);
  },

  apiCollection: async ({ api, workspace }, use) => {
    const factory = new ApiCollectionFactory(api);
    const apiCollection = await factory.create({ workspaceId: workspace.id });
    await use(apiCollection);
  },

  authenticatedUser: async ({ api }, use, testInfo) => {
    const created = await createUserWithStorage(api, testInfo, "auth");
    await use(created);
  },

  adminUser: async ({ api }, use, testInfo) => {
    const created = await createUserWithStorage(api, testInfo, "admin");
    await use(created);
  },

  qaUser: async ({ api }, use, testInfo) => {
    const created = await createUserWithStorage(api, testInfo, "qa");
    await use(created);
  },

  viewerUser: async ({ api }, use, testInfo) => {
    const created = await createUserWithStorage(api, testInfo, "viewer");
    await use(created);
  },

  workspaceOwner: async ({ api }, use, testInfo) => {
    const created = await createUserWithStorage(api, testInfo, "owner");
    await use(created);
  },
});

export { expect };
