import { ApiHelper } from "../helpers/api";
import { uniqueId } from "../helpers/random";
import { TestPlan } from "../types";

export interface CreateTestPlanInput {
  projectId: string;
  workspaceId: string;
  testCaseIds?: string[];
  testPlan?: Partial<TestPlan>;
}

export class TestPlanFactory {
  constructor(private api: ApiHelper) {}

  build(input: CreateTestPlanInput) {
    return {
      project_id: input.projectId,
      workspace_id: input.workspaceId,
      name: input.testPlan?.name || `Test Plan ${uniqueId()}`,
      description: input.testPlan?.description || "Auto-generated test plan",
      test_case_ids: input.testCaseIds || [],
    };
  }

  async create(input: CreateTestPlanInput): Promise<TestPlan> {
    return this.api.createTestPlan(this.build(input));
  }
}
