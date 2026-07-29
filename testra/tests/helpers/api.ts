import { APIRequestContext, APIResponse } from "@playwright/test";
import {
  ApiCollection,
  ApiEnvironment,
  ApiRequest,
  AuthResult,
  AutomationProject,
  Defect,
  Organization,
  Project,
  TestCase,
  TestPlan,
  TestRun,
  User,
  Workspace,
} from "../types";

const API_URL = process.env.TEST_API_URL || "http://localhost:8080";

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
  }
}

interface Envelope<T> {
  data?: T;
  meta?: Record<string, unknown>;
  error?: { code: string; message: string };
}

export class ApiHelper {
  private csrfToken: string | null = null;

  constructor(
    public request: APIRequestContext,
    public baseURL = API_URL,
  ) {}

  async unwrap<T>(res: APIResponse): Promise<T> {
    const body = (await res.json().catch(() => ({}))) as Envelope<T>;
    if (!res.ok()) {
      const code = body.error?.code || "UNKNOWN";
      const msg = body.error?.message || `Request failed with ${res.status()}`;
      throw new ApiError(res.status(), code, `${code}: ${msg}`);
    }
    const data = body.data as any;
    if (data && typeof data === "object" && Object.prototype.hasOwnProperty.call(data, "data")) {
      return data.data as T;
    }
    return body.data as T;
  }

  async unwrapEnvelope<T, M = Record<string, unknown>>(res: APIResponse): Promise<{ data: T; meta?: M }> {
    const body = (await res.json().catch(() => ({}))) as Envelope<T>;
    if (!res.ok()) {
      const code = body.error?.code || "UNKNOWN";
      const msg = body.error?.message || `Request failed with ${res.status()}`;
      throw new ApiError(res.status(), code, `${code}: ${msg}`);
    }
    return { data: body.data as T, meta: body.meta as unknown as M | undefined };
  }

  private isMutating(method: string): boolean {
    return !["GET", "HEAD", "OPTIONS"].includes(method.toUpperCase());
  }

  async ensureCsrf(): Promise<string | null> {
    if (this.csrfToken) return this.csrfToken;
    const res = await this.request.get(`${this.baseURL}/api/v1/auth/csrf`);
    const body = (await res.json().catch(() => ({}))) as Envelope<{ csrf_token: string }>;
    this.csrfToken = body.data?.csrf_token || null;
    return this.csrfToken;
  }

  async requestWithCsrf<T>(
    method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE",
    path: string,
    data?: unknown,
    params?: Record<string, string | number | boolean | undefined>,
  ): Promise<T> {
    const url = `${this.baseURL}${path}`;
    const headers: Record<string, string> = {};
    if (this.isMutating(method)) {
      const token = await this.ensureCsrf();
      if (token) headers["X-CSRF-Token"] = token;
    }

    const cleanParams = params
      ? (Object.fromEntries(
          Object.entries(params).filter(([, v]) => v !== undefined && v !== null && v !== ""),
        ) as Record<string, string | number | boolean>)
      : undefined;

    let res: APIResponse;
    switch (method) {
      case "GET":
        res = await this.request.get(url, { headers, params: cleanParams });
        break;
      case "POST":
        res = await this.request.post(url, { headers, params: cleanParams, data });
        break;
      case "PUT":
        res = await this.request.put(url, { headers, params: cleanParams, data });
        break;
      case "PATCH":
        res = await this.request.patch(url, { headers, params: cleanParams, data });
        break;
      case "DELETE":
        res = await this.request.delete(url, { headers, params: cleanParams });
        break;
      default:
        throw new Error(`Unsupported method ${method}`);
    }
    return this.unwrap<T>(res);
  }

  async requestWithCsrfEnvelope<T, M = Record<string, unknown>>(
    method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE",
    path: string,
    data?: unknown,
    params?: Record<string, string | number | boolean | undefined>,
  ): Promise<{ data: T; meta?: M }> {
    const url = `${this.baseURL}${path}`;
    const headers: Record<string, string> = {};
    if (this.isMutating(method)) {
      const token = await this.ensureCsrf();
      if (token) headers["X-CSRF-Token"] = token;
    }

    const cleanParams = params
      ? (Object.fromEntries(
          Object.entries(params).filter(([, v]) => v !== undefined && v !== null && v !== ""),
        ) as Record<string, string | number | boolean>)
      : undefined;

    let res: APIResponse;
    switch (method) {
      case "GET":
        res = await this.request.get(url, { headers, params: cleanParams });
        break;
      case "POST":
        res = await this.request.post(url, { headers, params: cleanParams, data });
        break;
      case "PUT":
        res = await this.request.put(url, { headers, params: cleanParams, data });
        break;
      case "PATCH":
        res = await this.request.patch(url, { headers, params: cleanParams, data });
        break;
      case "DELETE":
        res = await this.request.delete(url, { headers, params: cleanParams });
        break;
      default:
        throw new Error(`Unsupported method ${method}`);
    }
    return this.unwrapEnvelope<T, M>(res);
  }

  async register(input: { email: string; password: string; name: string }): Promise<AuthResult> {
    return this.requestWithCsrf("POST", "/api/v1/auth/register", input);
  }

  async login(input: { email: string; password: string; mfa_code?: string }): Promise<AuthResult> {
    return this.requestWithCsrf("POST", "/api/v1/auth/login", input);
  }

  async logout(): Promise<void> {
    await this.requestWithCsrf("POST", "/api/v1/auth/logout", {});
  }

  async me(): Promise<User> {
    return this.requestWithCsrf("GET", "/api/v1/auth/me");
  }

  async createOrganization(input: { name: string; slug: string; description?: string }): Promise<Organization> {
    return this.requestWithCsrf("POST", "/api/v1/organizations", input);
  }

  async listOrganizations(): Promise<Organization[]> {
    return this.requestWithCsrf("GET", "/api/v1/organizations");
  }

  async createWorkspace(input: {
    organization_id: string;
    name: string;
    slug: string;
    description?: string;
  }): Promise<Workspace> {
    return this.requestWithCsrf("POST", "/api/v1/workspaces", input);
  }

  async listWorkspaces(organizationId: string): Promise<Workspace[]> {
    return this.requestWithCsrf("GET", "/api/v1/workspaces", undefined, { organization_id: organizationId });
  }

  async createProject(input: { workspace_id: string; name: string; key: string; description?: string }): Promise<Project> {
    return this.requestWithCsrf("POST", "/api/v1/projects", input);
  }

  async listProjects(workspaceId: string): Promise<Project[]> {
    return this.requestWithCsrf("GET", "/api/v1/projects", undefined, { workspace_id: workspaceId });
  }

  async createTestCase(input: {
    workspace_id: string;
    project_id: string;
    title: string;
    description?: string;
    preconditions?: string;
    status?: string;
    priority?: string;
    tags?: string[];
    steps?: Array<{ action: string; expected: string; test_data?: string }>;
  }): Promise<TestCase> {
    return this.requestWithCsrf("POST", "/api/v1/test-cases", input);
  }

  async listTestCases(projectId: string): Promise<{ data: TestCase[]; meta?: { has_more?: boolean } }> {
    return this.requestWithCsrfEnvelope("GET", "/api/v1/test-cases", undefined, { project_id: projectId });
  }

  async getTestCase(id: string): Promise<TestCase> {
    return this.requestWithCsrf("GET", `/api/v1/test-cases/${id}`);
  }

  async updateTestCase(id: string, input: Partial<TestCase>): Promise<TestCase> {
    return this.requestWithCsrf("PUT", `/api/v1/test-cases/${id}`, input);
  }

  async deleteTestCase(id: string): Promise<void> {
    await this.requestWithCsrf("DELETE", `/api/v1/test-cases/${id}`);
  }

  async createTestPlan(input: {
    project_id: string;
    workspace_id: string;
    name: string;
    description?: string;
    test_case_ids?: string[];
  }): Promise<TestPlan> {
    return this.requestWithCsrf("POST", "/api/v1/test-plans", input);
  }

  async getTestPlan(id: string): Promise<TestPlan> {
    return this.requestWithCsrf("GET", `/api/v1/test-plans/${id}`);
  }

  async deleteTestPlan(id: string): Promise<void> {
    await this.requestWithCsrf("DELETE", `/api/v1/test-plans/${id}`);
  }

  async createTestRun(input: {
    project_id: string;
    workspace_id: string;
    name: string;
    test_case_ids?: string[];
    source?: string;
  }): Promise<TestRun> {
    return this.requestWithCsrf("POST", "/api/v1/test-runs", input);
  }

  async getTestRun(id: string): Promise<TestRun> {
    return this.requestWithCsrf("GET", `/api/v1/test-runs/${id}`);
  }

  async deleteTestRun(id: string): Promise<void> {
    await this.requestWithCsrf("DELETE", `/api/v1/test-runs/${id}`);
  }

  async createDefect(input: {
    workspace_id: string;
    project_id: string;
    title: string;
    description?: string;
    severity?: string;
    priority?: string;
  }): Promise<Defect> {
    return this.requestWithCsrf("POST", "/api/v1/defects", input);
  }

  async listDefects(projectId: string): Promise<{ data: Defect[]; meta?: { has_more?: boolean } }> {
    return this.requestWithCsrfEnvelope("GET", "/api/v1/defects", undefined, { project_id: projectId });
  }

  async createAutomationProject(input: {
    workspace_id: string;
    project_id?: string;
    name: string;
    framework: string;
    repository_url?: string;
    branch?: string;
    command?: string;
  }): Promise<AutomationProject> {
    return this.requestWithCsrf("POST", "/api/v1/automation/projects", input);
  }

  async listAutomationProjects(workspaceId: string): Promise<{ data: AutomationProject[]; meta?: { has_more?: boolean } }> {
    return this.requestWithCsrfEnvelope("GET", "/api/v1/automation/projects", undefined, { workspace_id: workspaceId });
  }

  async createApiCollection(input: { workspace_id: string; name: string; description?: string }): Promise<ApiCollection> {
    return this.requestWithCsrf("POST", "/api/v1/api-collections", input);
  }

  async createApiEnvironment(input: { workspace_id: string; name: string; variables?: Record<string, string> }): Promise<ApiEnvironment> {
    return this.requestWithCsrf("POST", "/api/v1/api-environments", input);
  }

  async createApiRequest(input: {
    collection_id: string;
    name: string;
    method: string;
    url: string;
    headers?: Record<string, string>;
    body?: string;
  }): Promise<ApiRequest> {
    return this.requestWithCsrf("POST", "/api/v1/api-requests", input);
  }

  async executeApiRequest(input: {
    workspace_id: string;
    request_id: string;
    environment_id?: string;
  }): Promise<unknown> {
    return this.requestWithCsrf("POST", "/api/v1/api-executions", input);
  }

  async search(workspaceId: string, query: string, limit = 10): Promise<{ data: { results: unknown[] } }> {
    return this.requestWithCsrfEnvelope("GET", "/api/v1/search", undefined, { workspace_id: workspaceId, q: query, limit });
  }

  async createNotification(input: { workspace_id: string; title: string; message: string }): Promise<unknown> {
    return this.requestWithCsrf("POST", "/api/v1/notifications", input);
  }

  async listNotifications(): Promise<{ data: unknown[] }> {
    return this.requestWithCsrfEnvelope("GET", "/api/v1/notifications");
  }

  async createIntegration(input: { workspace_id: string; name: string; provider: string; config?: unknown }): Promise<unknown> {
    return this.requestWithCsrf("POST", "/api/v1/integrations", input);
  }

  async listIntegrations(workspaceId: string): Promise<{ data: unknown[] }> {
    return this.requestWithCsrfEnvelope("GET", "/api/v1/integrations", undefined, { workspace_id: workspaceId });
  }

  async analyticsSummary(workspaceId: string): Promise<unknown> {
    return this.requestWithCsrf("GET", "/api/v1/analytics/summary", undefined, { workspace_id: workspaceId });
  }

  async analyticsTrends(workspaceId: string, params?: { project_id?: string; start?: string; end?: string }): Promise<unknown> {
    return this.requestWithCsrf("GET", "/api/v1/analytics/trends", undefined, {
      workspace_id: workspaceId,
      project_id: params?.project_id,
      start: params?.start,
      end: params?.end,
    });
  }

  async analyticsMetrics(workspaceId: string, params?: Record<string, string | number | boolean | undefined>): Promise<unknown> {
    return this.requestWithCsrf("GET", "/api/v1/analytics/metrics", undefined, {
      workspace_id: workspaceId,
      ...params,
    });
  }

  async analyticsRecentActivity(workspaceId: string, params?: Record<string, string | number | boolean | undefined>): Promise<unknown> {
    return this.requestWithCsrf("GET", "/api/v1/analytics/recent-activity", undefined, {
      workspace_id: workspaceId,
      ...params,
    });
  }

  async exportMetricsCSV(workspaceId: string, params?: Record<string, string | number | boolean | undefined>): Promise<string> {
    const url = `${this.baseURL}/api/v1/analytics/export/csv`;
    const cleanParams: Record<string, string | number | boolean> = { workspace_id: workspaceId };
    if (params) {
      for (const [k, v] of Object.entries(params)) {
        if (v !== undefined && v !== null && v !== "") {
          cleanParams[k] = v;
        }
      }
    }
    const res = await this.request.get(url, { params: cleanParams });
    if (!res.ok()) {
      throw new ApiError(res.status(), "CSV_EXPORT_FAILED", `Export failed with ${res.status()}`);
    }
    return await res.text();
  }

  async refreshToken(refreshToken: string): Promise<AuthResult> {
    return this.requestWithCsrf("POST", "/api/v1/auth/refresh", { refresh_token: refreshToken });
  }

  async getCsrfToken(): Promise<string> {
    const res = await this.request.get(`${this.baseURL}/api/v1/auth/csrf`);
    const body = (await res.json().catch(() => ({}))) as Envelope<{ csrf_token: string }>;
    this.csrfToken = body.data?.csrf_token || null;
    return this.csrfToken || "";
  }

  async requestPasswordReset(email: string): Promise<void> {
    await this.requestWithCsrf("POST", "/api/v1/auth/password-reset/request", { email });
  }

  async resetPassword(token: string, newPassword: string): Promise<void> {
    await this.requestWithCsrf("POST", "/api/v1/auth/password-reset/confirm", {
      token,
      new_password: newPassword,
    });
  }

  async listTestCasesPaginated(projectId: string, params?: { cursor?: string; limit?: number }): Promise<{ data: TestCase[]; meta: { next_cursor?: string; has_more: boolean } }> {
    return this.requestWithCsrfEnvelope<TestCase[], { next_cursor?: string; has_more: boolean }>("GET", "/api/v1/test-cases", undefined, {
      project_id: projectId,
      cursor: params?.cursor,
      limit: params?.limit,
    }) as Promise<{ data: TestCase[]; meta: { next_cursor?: string; has_more: boolean } }>;
  }

  async listDefectsPaginated(projectId: string, params?: { cursor?: string; limit?: number }): Promise<{ data: Defect[]; meta: { next_cursor?: string; has_more: boolean } }> {
    return this.requestWithCsrfEnvelope<Defect[], { next_cursor?: string; has_more: boolean }>("GET", "/api/v1/defects", undefined, {
      project_id: projectId,
      cursor: params?.cursor,
      limit: params?.limit,
    }) as Promise<{ data: Defect[]; meta: { next_cursor?: string; has_more: boolean } }>;
  }

  async getProject(id: string): Promise<Project> {
    return this.requestWithCsrf("GET", `/api/v1/projects/${id}`);
  }

  async updateProject(id: string, input: Partial<Project>): Promise<Project> {
    return this.requestWithCsrf("PUT", `/api/v1/projects/${id}`, input);
  }

  async deleteProject(id: string): Promise<void> {
    await this.requestWithCsrf("DELETE", `/api/v1/projects/${id}`);
  }

  async getDefect(id: string): Promise<Defect> {
    return this.requestWithCsrf("GET", `/api/v1/defects/${id}`);
  }

  async updateDefect(id: string, input: Partial<Defect>): Promise<Defect> {
    return this.requestWithCsrf("PUT", `/api/v1/defects/${id}`, input);
  }

  async deleteDefect(id: string): Promise<void> {
    await this.requestWithCsrf("DELETE", `/api/v1/defects/${id}`);
  }

  async getAutomationProject(id: string): Promise<AutomationProject> {
    return this.requestWithCsrf("GET", `/api/v1/automation/projects/${id}`);
  }

  async deleteAutomationProject(id: string): Promise<void> {
    await this.requestWithCsrf("DELETE", `/api/v1/automation/projects/${id}`);
  }

  async getApiCollection(id: string): Promise<ApiCollection> {
    return this.requestWithCsrf("GET", `/api/v1/api-collections/${id}`);
  }

  async deleteApiCollection(id: string): Promise<void> {
    await this.requestWithCsrf("DELETE", `/api/v1/api-collections/${id}`);
  }

  async getWorkspace(id: string): Promise<Workspace> {
    return this.requestWithCsrf("GET", `/api/v1/workspaces/${id}`);
  }

  async listWorkspacesPaginated(organizationId: string, params?: { cursor?: string; limit?: number }): Promise<{ data: Workspace[]; meta: { next_cursor?: string; has_more: boolean } }> {
    return this.requestWithCsrfEnvelope<Workspace[], { next_cursor?: string; has_more: boolean }>("GET", "/api/v1/workspaces", undefined, {
      organization_id: organizationId,
      cursor: params?.cursor,
      limit: params?.limit,
    }) as Promise<{ data: Workspace[]; meta: { next_cursor?: string; has_more: boolean } }>;
  }

  async listOrganizationsPaginated(params?: { cursor?: string; limit?: number }): Promise<{ data: Organization[]; meta: { next_cursor?: string; has_more: boolean } }> {
    return this.requestWithCsrfEnvelope<Organization[], { next_cursor?: string; has_more: boolean }>("GET", "/api/v1/organizations", undefined, {
      cursor: params?.cursor,
      limit: params?.limit,
    }) as Promise<{ data: Organization[]; meta: { next_cursor?: string; has_more: boolean } }>;
  }

  async markNotificationRead(id: string): Promise<void> {
    await this.requestWithCsrf("PATCH", `/api/v1/notifications/${id}`, { read: true });
  }

  async deleteIntegration(id: string): Promise<void> {
    await this.requestWithCsrf("DELETE", `/api/v1/integrations/${id}`);
  }

  async getIntegration(id: string): Promise<unknown> {
    return this.requestWithCsrf("GET", `/api/v1/integrations/${id}`);
  }

  async searchCases(workspaceId: string, query: string, params?: { cursor?: string; limit?: number }): Promise<{ data: TestCase[]; meta: { next_cursor?: string; has_more: boolean } }> {
    return this.requestWithCsrfEnvelope<TestCase[], { next_cursor?: string; has_more: boolean }>("GET", "/api/v1/test-cases/search", undefined, {
      workspace_id: workspaceId,
      q: query,
      cursor: params?.cursor,
      limit: params?.limit,
    }) as Promise<{ data: TestCase[]; meta: { next_cursor?: string; has_more: boolean } }>;
  }

  async rawRequest(method: string, path: string, data?: unknown, headers?: Record<string, string>): Promise<APIResponse> {
    const url = `${this.baseURL}${path}`;
    const allHeaders: Record<string, string> = { ...headers };
    if (this.isMutating(method)) {
      const token = await this.ensureCsrf();
      if (token) allHeaders["X-CSRF-Token"] = token;
    }
    switch (method.toUpperCase()) {
      case "GET":
        return this.request.get(url, { headers: allHeaders });
      case "POST":
        return this.request.post(url, { headers: allHeaders, data });
      case "PUT":
        return this.request.put(url, { headers: allHeaders, data });
      case "PATCH":
        return this.request.patch(url, { headers: allHeaders, data });
      case "DELETE":
        return this.request.delete(url, { headers: allHeaders });
      default:
        throw new Error(`Unsupported method ${method}`);
    }
  }
}
