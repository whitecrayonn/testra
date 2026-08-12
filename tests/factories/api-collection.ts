import { ApiHelper } from "../helpers/api";
import { uniqueId } from "../helpers/random";
import { ApiCollection, ApiEnvironment, ApiRequest } from "../types";

export interface CreateApiCollectionInput {
  workspaceId: string;
  collection?: Partial<ApiCollection>;
  environment?: Partial<ApiEnvironment>;
  request?: Partial<ApiRequest>;
}

export class ApiCollectionFactory {
  constructor(private api: ApiHelper) {}

  async create(input: CreateApiCollectionInput): Promise<{
    collection: ApiCollection;
    environment: ApiEnvironment;
    request: ApiRequest;
  }> {
    const collection = await this.api.createApiCollection({
      workspace_id: input.workspaceId,
      name: input.collection?.name || `Collection ${uniqueId()}`,
      description: input.collection?.description,
    });

    const environment = await this.api.createApiEnvironment({
      workspace_id: input.workspaceId,
      name: input.environment?.name || `Environment ${uniqueId()}`,
      variables: input.environment?.variables || { baseUrl: "https://jsonplaceholder.typicode.com" },
    });

    const request = await this.api.createApiRequest({
      collection_id: collection.id,
      name: input.request?.name || `Request ${uniqueId()}`,
      method: input.request?.method || "GET",
      url: input.request?.url || "https://jsonplaceholder.typicode.com/posts/1",
      headers: input.request?.headers,
      body: input.request?.body,
    });

    return { collection, environment, request };
  }
}
