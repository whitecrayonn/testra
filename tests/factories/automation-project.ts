import { ApiHelper } from "../helpers/api";
import { uniqueId } from "../helpers/random";
import { AutomationProject } from "../types";

export interface CreateAutomationProjectInput {
  workspaceId: string;
  projectId?: string;
  automationProject?: Partial<AutomationProject>;
}

export class AutomationProjectFactory {
  constructor(private api: ApiHelper) {}

  build(input: CreateAutomationProjectInput) {
    return {
      workspace_id: input.workspaceId,
      project_id: input.projectId,
      name: input.automationProject?.name || `Automation Project ${uniqueId()}`,
      framework: input.automationProject?.framework || "junit",
      repository_url: input.automationProject?.repository_url,
      branch: input.automationProject?.branch,
      command: input.automationProject?.command,
    };
  }

  async create(input: CreateAutomationProjectInput): Promise<AutomationProject> {
    return this.api.createAutomationProject(this.build(input));
  }
}
