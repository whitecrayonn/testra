import { ApiHelper } from "../helpers/api";
import { uniqueId } from "../helpers/random";
import { TestCase } from "../types";

export interface CreateTestCaseInput {
  projectId: string;
  workspaceId: string;
  testCase?: Partial<TestCase>;
}

export class TestCaseFactory {
  constructor(private api: ApiHelper) {}

  build(input: CreateTestCaseInput) {
    return {
      workspace_id: input.workspaceId,
      project_id: input.projectId,
      title: input.testCase?.title || `Test Case ${uniqueId()}`,
      description: input.testCase?.description || "Auto-generated test case",
      preconditions: input.testCase?.preconditions,
      status: input.testCase?.status || "draft",
      priority: input.testCase?.priority || "medium",
      tags: input.testCase?.tags || ["e2e", "automated"],
      steps: input.testCase?.steps || [
        { action: "Open application", expected: "Login page visible" },
        { action: "Enter credentials", expected: "Dashboard loads" },
      ],
    };
  }

  async create(input: CreateTestCaseInput): Promise<TestCase> {
    return this.api.createTestCase(this.build(input));
  }
}
