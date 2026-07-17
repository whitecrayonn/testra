export interface Organization {
  id: string;
  name: string;
  slug: string;
  created_at: string;
  updated_at: string;
}

export interface Workspace {
  id: string;
  organization_id: string;
  name: string;
  slug: string;
  created_at: string;
  updated_at: string;
}

export interface Project {
  id: string;
  workspace_id: string;
  name: string;
  key: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface APIKey {
  id: string;
  workspace_id: string;
  name: string;
  prefix: string;
  scopes: string[];
  expires_at: string | null;
  created_at: string;
  last_used_at: string | null;
}

export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
}

export interface Member {
  id: string;
  organization_id: string;
  user_id: string;
  role: string;
  email?: string;
  name?: string;
  created_at: string;
}

export interface Role {
  id: string;
  name: string;
  description: string;
  permissions: string[];
}
