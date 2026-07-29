import { ApiHelper } from "../helpers/api";
import { uniqueId, slugify } from "../helpers/random";
import { Organization, Workspace } from "../types";

export interface CreateWorkspaceInput {
  ownerId?: string;
  organization?: Partial<Organization>;
  workspace?: Partial<Workspace>;
}

export class WorkspaceFactory {
  constructor(private api: ApiHelper) {}

  async create(input: CreateWorkspaceInput = {}): Promise<Workspace & { ownerPassword?: string; ownerEmail?: string }> {
    let org: Organization;
    if (!input.organization?.id) {
      const name = input.organization?.name || `Org ${uniqueId()}`;
      org = await this.api.createOrganization({
        name,
        slug: input.organization?.slug || slugify(name),
        description: input.organization?.description,
      });
    } else {
      org = input.organization as Organization;
    }

    const wsName = input.workspace?.name || `Workspace ${uniqueId()}`;
    return this.api.createWorkspace({
      organization_id: org.id,
      name: wsName,
      slug: input.workspace?.slug || slugify(wsName),
      description: input.workspace?.description,
    });
  }
}
