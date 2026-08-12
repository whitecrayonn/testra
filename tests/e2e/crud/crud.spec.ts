import { test, expect } from "../../fixtures";

test.describe("Test Case CRUD @crud @testcases @regression", () => {
  test("create test case via API", async ({ api, workspace, project }) => {
    const tc = await api.createTestCase({
      workspace_id: workspace.id,
      project_id: project.id,
      title: "CRUD Test Case - Create",
      description: "Created via API for CRUD testing",
      status: "draft",
      priority: "medium",
      tags: ["crud", "regression"],
      steps: [
        { action: "Step 1 action", expected: "Step 1 expected", test_data: "" },
        { action: "Step 2 action", expected: "Step 2 expected", test_data: "test data" },
      ],
    });

    expect(tc.id).toBeTruthy();
    expect(tc.title).toBe("CRUD Test Case - Create");
    expect(tc.status).toBe("draft");
  });

  test("read test case via API", async ({ api, workspace: _workspace, project: _project, testCase }) => {
    const fetched = await api.getTestCase(testCase.id);
    expect(fetched.id).toBe(testCase.id);
    expect(fetched.title).toBe(testCase.title);
  });

  test("update test case via API", async ({ api, workspace: _workspace, project: _project, testCase }) => {
    const updated = await api.updateTestCase(testCase.id, {
      title: "Updated Title - CRUD",
      description: "Updated description",
      status: "active",
      priority: "high",
      tags: ["updated", "crud"],
    } as any);

    expect(updated.title).toBe("Updated Title - CRUD");
    expect(updated.status).toBe("active");
    expect(updated.priority).toBe("high");
  });

  test("delete test case via API", async ({ api, workspace: _workspace, project: _project, testCase }) => {
    await api.deleteTestCase(testCase.id);
    await expect(api.getTestCase(testCase.id)).rejects.toThrow();
  });

  test("list test cases for project", async ({ api, workspace: _workspace, project, testCase }) => {
    const list = await api.listTestCases(project.id);
    expect(list.data.length).toBeGreaterThan(0);
    expect(list.data.some((tc: any) => tc.id === testCase.id)).toBeTruthy();
  });

  test("create test case with empty title fails", async ({ api, workspace, project }) => {
    await expect(
      api.createTestCase({
        workspace_id: workspace.id,
        project_id: project.id,
        title: "",
      }),
    ).rejects.toThrow();
  });

  test("create test case with invalid workspace fails", async ({ api, project }) => {
    await expect(
      api.createTestCase({
        workspace_id: "00000000-0000-0000-0000-000000000000",
        project_id: project.id,
        title: "Should Fail",
      }),
    ).rejects.toThrow();
  });

  test("create test case with invalid project fails", async ({ api, workspace }) => {
    await expect(
      api.createTestCase({
        workspace_id: workspace.id,
        project_id: "00000000-0000-0000-0000-000000000000",
        title: "Should Fail",
      }),
    ).rejects.toThrow();
  });

  test("get non-existent test case fails", async ({ api }) => {
    await expect(
      api.getTestCase("00000000-0000-0000-0000-000000000000"),
    ).rejects.toThrow();
  });

  test("update non-existent test case fails", async ({ api }) => {
    await expect(
      api.updateTestCase("00000000-0000-0000-0000-000000000000", { title: "Nope" } as any),
    ).rejects.toThrow();
  });

  test("delete non-existent test case fails", async ({ api }) => {
    await expect(
      api.deleteTestCase("00000000-0000-0000-0000-000000000000"),
    ).rejects.toThrow();
  });
});

test.describe("Project CRUD @crud @projects @regression", () => {
  test("create project via API", async ({ api, workspace }) => {
    const project = await api.createProject({
      workspace_id: workspace.id,
      name: "CRUD Project",
      key: "CRUD",
      description: "Created for CRUD testing",
    });
    expect(project.id).toBeTruthy();
    expect(project.name).toBe("CRUD Project");
  });

  test("read project via API", async ({ api, project }) => {
    const fetched = await api.getProject(project.id);
    expect(fetched.id).toBe(project.id);
    expect(fetched.name).toBe(project.name);
  });

  test("update project via API", async ({ api, project }) => {
    const updated = await api.updateProject(project.id, {
      name: "Updated Project Name",
      description: "Updated description",
    } as any);
    expect(updated.name).toBe("Updated Project Name");
  });

  test("delete project via API", async ({ api, workspace }) => {
    const project = await api.createProject({
      workspace_id: workspace.id,
      name: "Delete Me",
      key: "DELM",
    });
    await api.deleteProject(project.id);
    await expect(api.getProject(project.id)).rejects.toThrow();
  });

  test("list projects for workspace", async ({ api, workspace, project }) => {
    const list = await api.listProjects(workspace.id);
    expect(list.length).toBeGreaterThan(0);
    expect(list.some((p: any) => p.id === project.id)).toBeTruthy();
  });

  test("create project with empty name fails", async ({ api, workspace }) => {
    await expect(
      api.createProject({
        workspace_id: workspace.id,
        name: "",
        key: "FAIL",
      }),
    ).rejects.toThrow();
  });

  test("create project with empty key fails", async ({ api, workspace }) => {
    await expect(
      api.createProject({
        workspace_id: workspace.id,
        name: "Should Fail",
        key: "",
      }),
    ).rejects.toThrow();
  });

  test("create project with invalid workspace fails", async ({ api }) => {
    await expect(
      api.createProject({
        workspace_id: "00000000-0000-0000-0000-000000000000",
        name: "Should Fail",
        key: "FAIL",
      }),
    ).rejects.toThrow();
  });

  test("get non-existent project fails", async ({ api }) => {
    await expect(api.getProject("00000000-0000-0000-0000-000000000000")).rejects.toThrow();
  });
});

test.describe("Defect CRUD @crud @defects @regression", () => {
  test("create defect via API", async ({ api, workspace, project }) => {
    const defect = await api.createDefect({
      workspace_id: workspace.id,
      project_id: project.id,
      title: "CRUD Defect",
      description: "Created for CRUD testing",
      severity: "high",
      priority: "high",
    });
    expect(defect.id).toBeTruthy();
    expect(defect.title).toBe("CRUD Defect");
  });

  test("read defect via API", async ({ api, defect }) => {
    const fetched = await api.getDefect(defect.id);
    expect(fetched.id).toBe(defect.id);
    expect(fetched.title).toBe(defect.title);
  });

  test("update defect via API", async ({ api, defect }) => {
    const updated = await api.updateDefect(defect.id, {
      title: "Updated Defect Title",
      severity: "critical",
    } as any);
    expect(updated.title).toBe("Updated Defect Title");
  });

  test("delete defect via API", async ({ api, workspace, project }) => {
    const defect = await api.createDefect({
      workspace_id: workspace.id,
      project_id: project.id,
      title: "Delete Me Defect",
    });
    await api.deleteDefect(defect.id);
    await expect(api.getDefect(defect.id)).rejects.toThrow();
  });

  test("list defects for project", async ({ api, project, defect }) => {
    const list = await api.listDefects(project.id);
    expect(list.data.length).toBeGreaterThan(0);
    expect(list.data.some((d: any) => d.id === defect.id)).toBeTruthy();
  });

  test("create defect without project fails", async ({ api, workspace }) => {
    await expect(
      api.createDefect({
        workspace_id: workspace.id,
        project_id: "",
        title: "Should Fail",
      }),
    ).rejects.toThrow();
  });

  test("create defect with empty title fails", async ({ api, workspace, project }) => {
    await expect(
      api.createDefect({
        workspace_id: workspace.id,
        project_id: project.id,
        title: "",
      }),
    ).rejects.toThrow();
  });
});

test.describe("Automation Project CRUD @crud @automation @regression", () => {
  test("create automation project via API", async ({ api, workspace }) => {
    const auto = await api.createAutomationProject({
      workspace_id: workspace.id,
      name: "CRUD Automation",
      framework: "playwright",
      repository_url: "https://github.com/testra/crud-automation",
      branch: "main",
      command: "npx playwright test",
    });
    expect(auto.id).toBeTruthy();
    expect(auto.name).toBe("CRUD Automation");
  });

  test("read automation project via API", async ({ api, automationProject }) => {
    const fetched = await api.getAutomationProject(automationProject.id);
    expect(fetched.id).toBe(automationProject.id);
  });

  test("delete automation project via API", async ({ api, workspace }) => {
    const auto = await api.createAutomationProject({
      workspace_id: workspace.id,
      name: "Delete Me Automation",
      framework: "cypress",
    });
    await api.deleteAutomationProject(auto.id);
    await expect(api.getAutomationProject(auto.id)).rejects.toThrow();
  });

  test("create automation project with unsupported framework fails", async ({ api, workspace }) => {
    await expect(
      api.createAutomationProject({
        workspace_id: workspace.id,
        name: "Bad Framework",
        framework: "unsupported_framework_xyz",
      }),
    ).rejects.toThrow();
  });

  test("list automation projects for workspace", async ({ api, workspace, automationProject }) => {
    const list = await api.listAutomationProjects(workspace.id);
    expect(list.data.length).toBeGreaterThan(0);
    expect(list.data.some((a: any) => a.id === automationProject.id)).toBeTruthy();
  });
});

test.describe("API Collection CRUD @crud @api-testing @regression", () => {
  test("create api collection via API", async ({ api, workspace }) => {
    const collection = await api.createApiCollection({
      workspace_id: workspace.id,
      name: "CRUD Collection",
      description: "Created for CRUD testing",
    });
    expect(collection.id).toBeTruthy();
    expect(collection.name).toBe("CRUD Collection");
  });

  test("read api collection via API", async ({ api, apiCollection }) => {
    const fetched = await api.getApiCollection(apiCollection.collection.id);
    expect(fetched.id).toBe(apiCollection.collection.id);
  });

  test("delete api collection via API", async ({ api, workspace }) => {
    const collection = await api.createApiCollection({
      workspace_id: workspace.id,
      name: "Delete Me Collection",
    });
    await api.deleteApiCollection(collection.id);
    await expect(api.getApiCollection(collection.id)).rejects.toThrow();
  });

  test("create api collection without workspace fails", async ({ api }) => {
    await expect(
      api.createApiCollection({
        workspace_id: "",
        name: "Should Fail",
      }),
    ).rejects.toThrow();
  });
});

test.describe("Test Plan CRUD @crud @testplans @regression", () => {
  test("create test plan via API", async ({ api, workspace, project, testCase }) => {
    const plan = await api.createTestPlan({
      workspace_id: workspace.id,
      project_id: project.id,
      name: "CRUD Test Plan",
      description: "Created for CRUD testing",
      test_case_ids: [testCase.id],
    });
    expect(plan.id).toBeTruthy();
    expect(plan.name).toBe("CRUD Test Plan");
  });

  test("read test plan via API", async ({ api, testPlan }) => {
    const fetched = await api.getTestPlan(testPlan.id);
    expect(fetched.id).toBe(testPlan.id);
    expect(fetched.name).toBe(testPlan.name);
  });

  test("delete test plan via API", async ({ api, workspace, project }) => {
    const plan = await api.createTestPlan({
      workspace_id: workspace.id,
      project_id: project.id,
      name: "Delete Me Plan",
    });
    await api.deleteTestPlan(plan.id);
    await expect(api.getTestPlan(plan.id)).rejects.toThrow();
  });

  test("create test plan with invalid test case IDs fails", async ({ api, workspace, project }) => {
    await expect(
      api.createTestPlan({
        workspace_id: workspace.id,
        project_id: project.id,
        name: "Bad Plan",
        test_case_ids: ["00000000-0000-0000-0000-000000000000"],
      }),
    ).rejects.toThrow();
  });
});

test.describe("Test Run CRUD @crud @testruns @regression", () => {
  test("create test run via API", async ({ api, workspace, project, testCase }) => {
    const run = await api.createTestRun({
      workspace_id: workspace.id,
      project_id: project.id,
      name: "CRUD Test Run",
      test_case_ids: [testCase.id],
      source: "manual",
    });
    expect(run.id).toBeTruthy();
    expect(run.name).toBe("CRUD Test Run");
  });

  test("read test run via API", async ({ api, testRun }) => {
    const fetched = await api.getTestRun(testRun.id);
    expect(fetched.id).toBe(testRun.id);
  });

  test("delete test run via API", async ({ api, workspace, project, testCase }) => {
    const run = await api.createTestRun({
      workspace_id: workspace.id,
      project_id: project.id,
      name: "Delete Me Run",
      test_case_ids: [testCase.id],
    });
    await api.deleteTestRun(run.id);
    await expect(api.getTestRun(run.id)).rejects.toThrow();
  });

  test("create test run without test cases fails", async ({ api, workspace, project }) => {
    await expect(
      api.createTestRun({
        workspace_id: workspace.id,
        project_id: project.id,
        name: "Empty Run",
        test_case_ids: [],
      }),
    ).rejects.toThrow();
  });
});

test.describe("Workspace CRUD @crud @workspaces @regression", () => {
  test("create workspace via API", async ({ api, user: _user }) => {
    const org = await api.createOrganization({
      name: "CRUD Org",
      slug: "crud-org-" + Date.now(),
    });
    const ws = await api.createWorkspace({
      organization_id: org.id,
      name: "CRUD Workspace",
      slug: "crud-ws-" + Date.now(),
    });
    expect(ws.id).toBeTruthy();
    expect(ws.name).toBe("CRUD Workspace");
  });

  test("read workspace via API", async ({ api, workspace }) => {
    const fetched = await api.getWorkspace(workspace.id);
    expect(fetched.id).toBe(workspace.id);
    expect(fetched.name).toBe(workspace.name);
  });

  test("create workspace with duplicate slug fails", async ({ api, workspace }) => {
    await expect(
      api.createWorkspace({
        organization_id: workspace.organization_id,
        name: "Duplicate Slug",
        slug: workspace.slug,
      }),
    ).rejects.toThrow();
  });

  test("create workspace with empty name fails", async ({ api }) => {
    await expect(
      api.createWorkspace({
        organization_id: "00000000-0000-0000-0000-000000000000",
        name: "",
        slug: "test",
      }),
    ).rejects.toThrow();
  });
});
