import { ApiHelper } from "../helpers/api";
import { uniqueId } from "../helpers/random";
import { Project } from "../types";

export interface CreateProjectInput {
  workspaceId: string;
  project?: Partial<Project>;
}

export class ProjectFactory {
  constructor(private api: ApiHelper) {}

  build(input: CreateProjectInput) {
    const name = input.project?.name || `Project ${uniqueId()}`;
    const key =
      input.project?.key ||
      name
        .toUpperCase()
        .replace(/[^A-Z0-9]/g, "")
        .slice(0, 10) || "PROJECT";
    return {
      workspace_id: input.workspaceId,
      name,
      key,
      description: input.project?.description,
    };
  }

  async create(input: CreateProjectInput): Promise<Project> {
    return this.api.createProject(this.build(input));
  }
}
