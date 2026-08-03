import { ApiHelper } from "../helpers/api";
import { uniqueId } from "../helpers/random";
import { Defect } from "../types";

export interface CreateDefectInput {
  projectId: string;
  workspaceId: string;
  defect?: Partial<Defect>;
}

export class DefectFactory {
  constructor(private api: ApiHelper) {}

  build(input: CreateDefectInput) {
    return {
      workspace_id: input.workspaceId,
      project_id: input.projectId,
      title: input.defect?.title || `Defect ${uniqueId()}`,
      description: input.defect?.description || "Auto-generated defect",
      severity: input.defect?.severity || "medium",
      priority: input.defect?.priority || "medium",
    };
  }

  async create(input: CreateDefectInput): Promise<Defect> {
    return this.api.createDefect(this.build(input));
  }
}
