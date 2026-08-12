import { ApiHelper } from "../helpers/api";
import { uniqueId } from "../helpers/random";
import { TestRun } from "../types";

export interface CreateTestRunInput {
  projectId: string;
  workspaceId: string;
  testCaseIds?: string[];
  testRun?: Partial<TestRun>;
}

export class TestRunFactory {
  constructor(private api: ApiHelper) {}

  build(input: CreateTestRunInput) {
    return {
      project_id: input.projectId,
      workspace_id: input.workspaceId,
      name: input.testRun?.name || `Test Run ${uniqueId()}`,
      test_case_ids: input.testCaseIds || [],
      source: input.testRun?.source || "manual",
    };
  }

  async create(input: CreateTestRunInput): Promise<TestRun> {
    return this.api.createTestRun(this.build(input));
  }
}
