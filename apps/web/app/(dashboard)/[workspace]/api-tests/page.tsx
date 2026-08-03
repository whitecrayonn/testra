"use client";

import { useEffect, useMemo, useCallback, useState } from "react";
import {
  Send,
  Save,
  Trash2,
  Plus,
  Folder,
  FileText,
  Globe,
  History,
  ChevronRight,
  ChevronDown,
  X,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { Switch } from "@/components/ui/switch";
import {
  listCollections,
  createCollection,
  deleteCollection,
  listFolders,
  createFolder,
  listRequests,
  createRequest,
  updateRequest,
  deleteRequest,
  executeRequest,
  listEnvironments,
  createEnvironment,
  updateEnvironment,
  deleteEnvironment,
  listRequestHistory,
} from "@/features/apitesting/api";
import type {
  APICollection,
  APIFolder,
  APIRequest,
  APIEnvironment,
  APIRequestHistory,
  KeyValuePair,
  ExecutionResponse,
  BodyType,
  AuthType,
  AuthConfig,
  HTTPMethod,
} from "@/types/apitesting";

type ViewMode = "request" | "environment" | "history";

const METHODS: HTTPMethod[] = ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"];
const AUTH_TYPES: AuthType[] = ["none", "bearer", "basic", "api_key"];
const BODY_TYPES: BodyType[] = ["none", "json", "raw", "form", "urlencoded"];

const emptyPair = (): KeyValuePair => ({ key: "", value: "", enabled: true });

const defaultRequest = (collectionId: string, workspaceId: string): APIRequest => ({
  id: "",
  workspace_id: workspaceId,
  collection_id: collectionId,
  folder_id: null,
  environment_id: null,
  name: "New Request",
  method: "GET",
  url: "",
  headers: [],
  query_params: [],
  auth_type: "none",
  auth_config: {},
  body_type: "none",
  body_content: "",
  variables: [],
  created_by: "",
  created_at: "",
  updated_at: "",
});


export default function APITestingPage() {
  const workspaceId =
    typeof window !== "undefined" ? localStorage.getItem("testra_workspace_id") || "" : "";

  const [collections, setCollections] = useState<APICollection[]>([]);
  const [folders, setFolders] = useState<APIFolder[]>([]);
  const [requests, setRequests] = useState<APIRequest[]>([]);
  const [environments, setEnvironments] = useState<APIEnvironment[]>([]);
  const [history, setHistory] = useState<APIRequestHistory[]>([]);

  const [selectedCollectionId, setSelectedCollectionId] = useState<string>("");
  const [selectedFolderId, setSelectedFolderId] = useState<string | null>(null);
  const [activeRequest, setActiveRequest] = useState<APIRequest | null>(null);
  const [activeEnvironment, setActiveEnvironment] = useState<APIEnvironment | null>(null);

  const [viewMode, setViewMode] = useState<ViewMode>("request");
  const [activeTab, setActiveTab] = useState<"params" | "headers" | "auth" | "body" | "variables">("params");
  const [responseTab, setResponseTab] = useState<"body" | "headers">("body");
  const [envTab, setEnvTab] = useState<"list" | "editor">("list");

  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [response, setResponse] = useState<ExecutionResponse | null>(null);

  const [collectionSearch, setCollectionSearch] = useState("");
  const [newCollectionName, setNewCollectionName] = useState("");
  const [newFolderName, setNewFolderName] = useState("");
  const [newEnvName, setNewEnvName] = useState("");

  const [expandedCollections, setExpandedCollections] = useState<Set<string>>(new Set());

  const filteredCollections = useMemo(() => {
    const q = collectionSearch.toLowerCase();
    return collections.filter((c) => c.name.toLowerCase().includes(q));
  }, [collections, collectionSearch]);

  const loadData = useCallback(async () => {
    if (!workspaceId) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const [collectionsRes, envsRes] = await Promise.all([
        listCollections(workspaceId),
        listEnvironments(workspaceId),
      ]);
      setCollections(collectionsRes.data);
      setEnvironments(envsRes.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load API testing data");
    } finally {
      setLoading(false);
    }
  }, [workspaceId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  async function loadCollectionContents(collectionId: string, folderId?: string | null) {
    try {
      const [foldersRes, requestsRes] = await Promise.all([
        listFolders(collectionId),
        listRequests(collectionId, folderId ? { folderId } : undefined),
      ]);
      setFolders(foldersRes.data);
      setRequests(requestsRes.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load collection contents");
    }
  }

  async function loadHistory(requestId: string) {
    try {
      const res = await listRequestHistory(requestId);
      setHistory(res.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load history");
    }
  }

  function toggleCollection(id: string) {
    setExpandedCollections((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
        setSelectedCollectionId("");
        setSelectedFolderId(null);
      } else {
        next.clear();
        next.add(id);
        setSelectedCollectionId(id);
        setSelectedFolderId(null);
        loadCollectionContents(id);
      }
      return next;
    });
  }

  function selectFolder(folderId: string) {
    setSelectedFolderId(folderId);
    if (selectedCollectionId) {
      listRequests(selectedCollectionId, { folderId }).then((res) => setRequests(res.data));
    }
  }

  function selectRequest(req: APIRequest) {
    setActiveRequest({ ...req });
    setViewMode("request");
    setActiveTab("params");
    setResponse(null);
    loadHistory(req.id);
  }

  function newRequest() {
    if (!selectedCollectionId) {
      setError("Select a collection first");
      return;
    }
    setActiveRequest(defaultRequest(selectedCollectionId, workspaceId));
    setViewMode("request");
    setActiveTab("params");
    setResponse(null);
  }

  async function createNewCollection() {
    if (!newCollectionName.trim() || !workspaceId) return;
    try {
      const c = await createCollection({
        workspace_id: workspaceId,
        name: newCollectionName.trim(),
      });
      setCollections((prev) => [c, ...prev]);
      setNewCollectionName("");
      setSelectedCollectionId(c.id);
      setExpandedCollections(new Set([c.id]));
      setFolders([]);
      setRequests([]);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create collection");
    }
  }

  async function createNewFolder() {
    if (!newFolderName.trim() || !selectedCollectionId || !workspaceId) return;
    try {
      const f = await createFolder({
        workspace_id: workspaceId,
        collection_id: selectedCollectionId,
        name: newFolderName.trim(),
      });
      setFolders((prev) => [f, ...prev]);
      setNewFolderName("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create folder");
    }
  }

  async function removeCollection(id: string) {
    try {
      await deleteCollection(id);
      setCollections((prev) => prev.filter((c) => c.id !== id));
      if (selectedCollectionId === id) {
        setSelectedCollectionId("");
        setFolders([]);
        setRequests([]);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete collection");
    }
  }

  async function createNewEnvironment() {
    if (!newEnvName.trim() || !workspaceId) return;
    try {
      const env = await createEnvironment({
        workspace_id: workspaceId,
        name: newEnvName.trim(),
      });
      setEnvironments((prev) => [env, ...prev]);
      setNewEnvName("");
      setActiveEnvironment(env);
      setEnvTab("editor");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create environment");
    }
  }

  async function saveEnvironment() {
    if (!activeEnvironment) return;
    try {
      if (activeEnvironment.id) {
        const updated = await updateEnvironment(activeEnvironment.id, {
          name: activeEnvironment.name,
          variables: activeEnvironment.variables,
        });
        setEnvironments((prev) =>
          prev.map((e) => (e.id === updated.id ? updated : e)),
        );
        setActiveEnvironment(updated);
      } else {
        const created = await createEnvironment({
          workspace_id: workspaceId,
          name: activeEnvironment.name,
          variables: activeEnvironment.variables,
        });
        setEnvironments((prev) => [created, ...prev]);
        setActiveEnvironment(created);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save environment");
    }
  }

  async function removeEnvironment(id: string) {
    try {
      await deleteEnvironment(id);
      setEnvironments((prev) => prev.filter((e) => e.id !== id));
      if (activeEnvironment?.id === id) {
        setActiveEnvironment(null);
        setEnvTab("list");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete environment");
    }
  }

  async function saveRequest() {
    if (!activeRequest) return;
    if (!activeRequest.name.trim() || !activeRequest.url.trim()) {
      setError("Request name and URL are required");
      return;
    }
    setSaving(true);
    try {
      const payload = {
        collection_id: activeRequest.collection_id,
        folder_id: activeRequest.folder_id || undefined,
        environment_id: activeRequest.environment_id || undefined,
        name: activeRequest.name,
        method: activeRequest.method,
        url: activeRequest.url,
        headers: activeRequest.headers,
        query_params: activeRequest.query_params,
        auth_type: activeRequest.auth_type,
        auth_config: activeRequest.auth_config,
        body_type: activeRequest.body_type,
        body_content: activeRequest.body_content,
        variables: activeRequest.variables,
      };

      if (activeRequest.id) {
        const updated = await updateRequest(activeRequest.id, payload);
        setActiveRequest(updated);
        setRequests((prev) => prev.map((r) => (r.id === updated.id ? updated : r)));
      } else {
        const created = await createRequest({
          workspace_id: workspaceId,
          ...payload,
        });
        setActiveRequest(created);
        setRequests((prev) => [created, ...prev]);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save request");
    } finally {
      setSaving(false);
    }
  }

  async function removeRequest() {
    if (!activeRequest?.id) return;
    try {
      await deleteRequest(activeRequest.id);
      setRequests((prev) => prev.filter((r) => r.id !== activeRequest.id));
      setActiveRequest(null);
      setResponse(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete request");
    }
  }

  async function sendRequest() {
    if (!activeRequest || !workspaceId) return;
    setSending(true);
    setResponse(null);
    try {
      const res = await executeRequest({
        workspace_id: workspaceId,
        request_id: activeRequest.id || undefined,
        environment_id: activeRequest.environment_id || undefined,
        request: activeRequest.id
          ? undefined
          : {
              method: activeRequest.method,
              url: activeRequest.url,
              headers: activeRequest.headers,
              query_params: activeRequest.query_params,
              auth_type: activeRequest.auth_type,
              auth_config: activeRequest.auth_config,
              body_type: activeRequest.body_type,
              body_content: activeRequest.body_content,
              variables: activeRequest.variables,
              environment_id: activeRequest.environment_id || undefined,
            },
        save: true,
      });
      setResponse(res);
      if (activeRequest.id) {
        loadHistory(activeRequest.id);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to execute request");
    } finally {
      setSending(false);
    }
  }

  function updatePair(
    field: "headers" | "query_params" | "variables",
    index: number,
    key: "key" | "value" | "enabled",
    value: string | boolean,
  ) {
    if (!activeRequest) return;
    const next = { ...activeRequest, [field]: [...activeRequest[field]] };
    (next[field] as KeyValuePair[])[index] = {
      ...next[field][index],
      [key]: value,
    };
    setActiveRequest(next);
  }

  function addPair(field: "headers" | "query_params" | "variables") {
    if (!activeRequest) return;
    setActiveRequest({
      ...activeRequest,
      [field]: [...activeRequest[field], emptyPair()],
    });
  }

  function removePair(field: "headers" | "query_params" | "variables", index: number) {
    if (!activeRequest) return;
    const next = [...activeRequest[field]];
    next.splice(index, 1);
    setActiveRequest({ ...activeRequest, [field]: next });
  }

  function updateEnvVariable(index: number, key: "key" | "value" | "enabled", value: string | boolean) {
    if (!activeEnvironment) return;
    const next = [...activeEnvironment.variables];
    next[index] = { ...next[index], [key]: value };
    setActiveEnvironment({ ...activeEnvironment, variables: next });
  }

  function addEnvVariable() {
    if (!activeEnvironment) return;
    setActiveEnvironment({
      ...activeEnvironment,
      variables: [...activeEnvironment.variables, emptyPair()],
    });
  }

  function removeEnvVariable(index: number) {
    if (!activeEnvironment) return;
    const next = [...activeEnvironment.variables];
    next.splice(index, 1);
    setActiveEnvironment({ ...activeEnvironment, variables: next });
  }

  function formatResponseBody(body: string) {
    try {
      return JSON.stringify(JSON.parse(body), null, 2);
    } catch {
      return body;
    }
  }

  if (!workspaceId) {
    return (
      <div className="space-y-6">
        <PageHeader title="API Testing" description="Build, run, and organize API tests." />
        <EmptyState
          icon={Globe}
          title="No workspace selected"
          description="Select a workspace from the dashboard to start testing APIs."
          action={{ label: "Go to Dashboard", href: "/dashboard" }}
        />
      </div>
    );
  }

  return (
    <div className="flex h-[calc(100vh-4rem)] flex-col gap-6">
      <PageHeader
        title="API Testing"
        description="Build, run, and organize API tests."
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant={viewMode === "request" ? "primary" : "secondary"}
              size="sm"
              onClick={() => setViewMode("request")}
            >
              Requests
            </Button>
            <Button
              variant={viewMode === "environment" ? "primary" : "secondary"}
              size="sm"
              onClick={() => setViewMode("environment")}
            >
              Environments
            </Button>
            <Button
              variant={viewMode === "history" ? "primary" : "secondary"}
              size="sm"
              onClick={() => setViewMode("history")}
            >
              History
            </Button>
          </div>
        }
      />

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
          <button
            onClick={() => setError(null)}
            className="float-right text-red-700 hover:text-red-900"
            aria-label="Dismiss"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      )}

      <div className="flex min-h-0 flex-1 gap-6">
        {/* Left sidebar */}
        <Card className="flex w-80 flex-col overflow-hidden">
          <div className="border-b border-slate-200 p-4">
            <div className="mb-3 flex items-center justify-between">
              <h2 className="font-semibold text-slate-900">Collections</h2>
              <Button variant="ghost" size="sm" onClick={newRequest} disabled={!selectedCollectionId}>
                <Plus className="h-4 w-4" />
              </Button>
            </div>
            <Input
              placeholder="Search collections..."
              value={collectionSearch}
              onChange={(e) => setCollectionSearch(e.target.value)}
              className="h-8 text-sm"
            />
            <div className="mt-2 flex gap-2">
              <Input
                placeholder="New collection"
                value={newCollectionName}
                onChange={(e) => setNewCollectionName(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && createNewCollection()}
                className="h-8 flex-1 text-sm"
              />
              <Button size="sm" variant="secondary" onClick={createNewCollection}>
                <Plus className="h-4 w-4" />
              </Button>
            </div>
          </div>

          <div className="flex-1 overflow-y-auto p-2">
            {loading && collections.length === 0 ? (
              <div className="space-y-2">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="h-8 animate-pulse rounded bg-slate-100" />
                ))}
              </div>
            ) : collections.length === 0 ? (
              <p className="p-4 text-center text-sm text-slate-500">No collections yet</p>
            ) : (
              <ul className="space-y-1">
                {filteredCollections.map((collection) => {
                  const expanded = expandedCollections.has(collection.id);
                  return (
                    <li key={collection.id}>
                      <button
                        onClick={() => toggleCollection(collection.id)}
                        className={`flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm ${
                          selectedCollectionId === collection.id
                            ? "bg-brand-50 text-brand-700"
                            : "text-slate-700 hover:bg-slate-100"
                        }`}
                      >
                        {expanded ? (
                          <ChevronDown className="h-3.5 w-3.5 text-slate-400" />
                        ) : (
                          <ChevronRight className="h-3.5 w-3.5 text-slate-400" />
                        )}
                        <Folder className="h-4 w-4 text-brand-500" />
                        <span className="flex-1 truncate">{collection.name}</span>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            removeCollection(collection.id);
                          }}
                          className="text-slate-400 hover:text-red-600"
                          title="Delete collection"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      </button>

                      {expanded && (
                        <div className="ml-4 mt-1 border-l border-slate-200 pl-2">
                          <div className="mb-2 flex items-center gap-1">
                            <Input
                              placeholder="New folder"
                              value={newFolderName}
                              onChange={(e) => setNewFolderName(e.target.value)}
                              onKeyDown={(e) => e.key === "Enter" && createNewFolder()}
                              className="h-7 flex-1 text-xs"
                            />
                            <Button variant="ghost" size="sm" className="h-7 px-1" onClick={createNewFolder}>
                              <Plus className="h-3.5 w-3.5" />
                            </Button>
                          </div>
                          {folders.map((folder) => (
                            <button
                              key={folder.id}
                              onClick={() => selectFolder(folder.id)}
                              className={`flex w-full items-center gap-2 rounded-md px-2 py-1 text-left text-sm ${
                                selectedFolderId === folder.id
                                  ? "bg-slate-100 text-slate-900"
                                  : "text-slate-600 hover:bg-slate-50"
                              }`}
                            >
                              <Folder className="h-3.5 w-3.5 text-slate-400" />
                              <span className="flex-1 truncate">{folder.name}</span>
                            </button>
                          ))}
                          {requests.map((req) => (
                            <button
                              key={req.id}
                              onClick={() => selectRequest(req)}
                              className={`flex w-full items-center gap-2 rounded-md px-2 py-1 text-left text-sm ${
                                activeRequest?.id === req.id
                                  ? "bg-slate-100 text-slate-900"
                                  : "text-slate-600 hover:bg-slate-50"
                              }`}
                            >
                              <FileText className="h-3.5 w-3.5 text-slate-400" />
                              <span className="flex-1 truncate">{req.name}</span>
                              <Badge variant="neutral" className="text-[10px]">
                                {req.method}
                              </Badge>
                            </button>
                          ))}
                        </div>
                      )}
                    </li>
                  );
                })}
              </ul>
            )}
          </div>
        </Card>

        {/* Main editor */}
        <div className="flex min-w-0 flex-1 flex-col">
          {viewMode === "environment" ? (
            <Card className="flex flex-1 flex-col overflow-hidden p-4">
              <div className="mb-4 flex items-center justify-between">
                <h2 className="font-semibold text-slate-900">Environments</h2>
                <div className="flex gap-2">
                  <Input
                    placeholder="New environment"
                    value={newEnvName}
                    onChange={(e) => setNewEnvName(e.target.value)}
                    onKeyDown={(e) => e.key === "Enter" && createNewEnvironment()}
                    className="h-8 w-48 text-sm"
                  />
                  <Button size="sm" variant="secondary" onClick={createNewEnvironment}>
                    <Plus className="mr-1 h-4 w-4" /> Add
                  </Button>
                </div>
              </div>

              {envTab === "list" ? (
                <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                  {environments.map((env) => (
                    <Card key={env.id} className="p-4">
                      <div className="flex items-start justify-between">
                        <div>
                          <h3 className="font-medium text-slate-900">{env.name}</h3>
                          <p className="text-sm text-slate-500">{env.variables.length} variables</p>
                        </div>
                        <div className="flex gap-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => {
                              setActiveEnvironment({ ...env });
                              setEnvTab("editor");
                            }}
                          >
                            Edit
                          </Button>
                          <Button variant="ghost" size="sm" onClick={() => removeEnvironment(env.id)}>
                            <Trash2 className="h-4 w-4 text-red-500" />
                          </Button>
                        </div>
                      </div>
                    </Card>
                  ))}
                  {environments.length === 0 && (
                    <EmptyState
                      icon={Globe}
                      title="No environments"
                      description="Create an environment to manage variables for your requests."
                      action={{ label: "Create Environment", onClick: () => setEnvTab("editor") }}
                    />
                  )}
                </div>
              ) : activeEnvironment ? (
                <div className="flex flex-1 flex-col gap-4">
                  <div className="flex items-center gap-2">
                    <Button variant="ghost" size="sm" onClick={() => setEnvTab("list")}>
                      Back
                    </Button>
                    <Input
                      value={activeEnvironment.name}
                      onChange={(e) => setActiveEnvironment({ ...activeEnvironment, name: e.target.value })}
                      className="max-w-md"
                    />
                    <Button size="sm" onClick={saveEnvironment} loading={saving}>
                      <Save className="mr-1 h-4 w-4" /> Save
                    </Button>
                  </div>
                  <div className="flex-1 overflow-y-auto">
                    <table className="w-full text-sm">
                      <thead className="text-left text-slate-500">
                        <tr>
                          <th className="pb-2 font-medium">Enabled</th>
                          <th className="pb-2 font-medium">Key</th>
                          <th className="pb-2 font-medium">Value</th>
                          <th className="pb-2 font-medium"></th>
                        </tr>
                      </thead>
                      <tbody>
                        {activeEnvironment.variables.map((v, i) => (
                          <tr key={i} className="border-b border-slate-100">
                            <td className="py-2 pr-2">
                              <Switch
                                checked={v.enabled}
                                onCheckedChange={(checked) => updateEnvVariable(i, "enabled", checked)}
                              />
                            </td>
                            <td className="py-2 pr-2">
                              <Input
                                value={v.key}
                                onChange={(e) => updateEnvVariable(i, "key", e.target.value)}
                                className="h-8"
                              />
                            </td>
                            <td className="py-2 pr-2">
                              <Input
                                value={v.value}
                                onChange={(e) => updateEnvVariable(i, "value", e.target.value)}
                                className="h-8"
                              />
                            </td>
                            <td className="py-2">
                              <Button variant="ghost" size="sm" onClick={() => removeEnvVariable(i)}>
                                <X className="h-4 w-4 text-red-500" />
                              </Button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                    <Button variant="secondary" size="sm" className="mt-4" onClick={addEnvVariable}>
                      <Plus className="mr-1 h-4 w-4" /> Add Variable
                    </Button>
                  </div>
                </div>
              ) : null}
            </Card>
          ) : viewMode === "history" ? (
            <Card className="flex flex-1 flex-col overflow-hidden p-4">
              <h2 className="mb-4 font-semibold text-slate-900">Execution History</h2>
              <div className="flex-1 overflow-y-auto space-y-2">
                {history.length === 0 ? (
                  <EmptyState
                    icon={History}
                    title="No history"
                    description="Run a request to start recording execution history."
                  />
                ) : (
                  history.map((h) => (
                    <Card key={h.id} className="p-3">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="font-medium text-slate-900">{h.name || h.url}</p>
                          <p className="text-xs text-slate-500">
                            {h.method} {h.url}
                          </p>
                        </div>
                        <div className="text-right text-sm">
                          <Badge
                            variant={h.response_status >= 400 || h.error ? "danger" : "success"}
                            className="mb-1"
                          >
                            {h.error ? "Error" : h.response_status || "—"}
                          </Badge>
                          <p className="text-xs text-slate-400">{h.response_time_ms} ms</p>
                        </div>
                      </div>
                    </Card>
                  ))
                )}
              </div>
            </Card>
          ) : activeRequest ? (
            <Card className="flex flex-1 flex-col overflow-hidden">
              <div className="border-b border-slate-200 p-4">
                <div className="mb-3 flex items-center gap-2">
                  <Input
                    value={activeRequest.name}
                    onChange={(e) => setActiveRequest({ ...activeRequest, name: e.target.value })}
                    className="max-w-md font-medium"
                    placeholder="Request name"
                  />
                  <select
                    value={activeRequest.method}
                    onChange={(e) => setActiveRequest({ ...activeRequest, method: e.target.value as HTTPMethod })}
                    className="h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm font-semibold text-slate-900 focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
                  >
                    {METHODS.map((m) => (
                      <option key={m} value={m}>
                        {m}
                      </option>
                    ))}
                  </select>
                  <Input
                    value={activeRequest.url}
                    onChange={(e) => setActiveRequest({ ...activeRequest, url: e.target.value })}
                    placeholder="https://api.example.com/resource"
                    className="flex-1"
                  />
                  <select
                    value={activeRequest.environment_id || ""}
                    onChange={(e) =>
                      setActiveRequest({
                        ...activeRequest,
                        environment_id: e.target.value || null,
                      })
                    }
                    className="h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-900 focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
                  >
                    <option value="">No environment</option>
                    {environments.map((env) => (
                      <option key={env.id} value={env.id}>
                        {env.name}
                      </option>
                    ))}
                  </select>
                  <Button onClick={sendRequest} loading={sending}>
                    <Send className="mr-2 h-4 w-4" /> Send
                  </Button>
                  <Button variant="secondary" onClick={saveRequest} loading={saving}>
                    <Save className="mr-2 h-4 w-4" /> Save
                  </Button>
                  {activeRequest.id && (
                    <Button variant="danger" onClick={removeRequest}>
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  )}
                </div>

                <div className="flex gap-1 border-b border-slate-200">
                  {(["params", "headers", "auth", "body", "variables"] as const).map((tab) => (
                    <button
                      key={tab}
                      onClick={() => setActiveTab(tab)}
                      className={`px-4 py-2 text-sm font-medium capitalize ${
                        activeTab === tab
                          ? "border-b-2 border-brand-500 text-brand-700"
                          : "text-slate-500 hover:text-slate-700"
                      }`}
                    >
                      {tab}
                    </button>
                  ))}
                </div>

                <div className="py-4">
                  {activeTab === "params" && (
                    <KeyValueEditor
                      pairs={activeRequest.query_params}
                      onChange={(index, key, value) => updatePair("query_params", index, key, value)}
                      onAdd={() => addPair("query_params")}
                      onRemove={(index) => removePair("query_params", index)}
                    />
                  )}
                  {activeTab === "headers" && (
                    <KeyValueEditor
                      pairs={activeRequest.headers}
                      onChange={(index, key, value) => updatePair("headers", index, key, value)}
                      onAdd={() => addPair("headers")}
                      onRemove={(index) => removePair("headers", index)}
                    />
                  )}
                  {activeTab === "auth" && (
                    <AuthEditor
                      authType={activeRequest.auth_type}
                      authConfig={activeRequest.auth_config}
                      onChange={(type, config) =>
                        setActiveRequest({ ...activeRequest, auth_type: type, auth_config: config })
                      }
                    />
                  )}
                  {activeTab === "body" && (
                    <BodyEditor
                      bodyType={activeRequest.body_type}
                      bodyContent={activeRequest.body_content}
                      onChange={(type, content) =>
                        setActiveRequest({ ...activeRequest, body_type: type, body_content: content })
                      }
                    />
                  )}
                  {activeTab === "variables" && (
                    <KeyValueEditor
                      pairs={activeRequest.variables}
                      onChange={(index, key, value) => updatePair("variables", index, key, value)}
                      onAdd={() => addPair("variables")}
                      onRemove={(index) => removePair("variables", index)}
                    />
                  )}
                </div>
              </div>

              <div className="flex-1 overflow-y-auto p-4">
                {response ? (
                  <div className="space-y-4">
                    <div className="flex items-center gap-4">
                      <Badge variant={response.result.error || response.result.status >= 400 ? "danger" : "success"}>
                        {response.result.status || "Error"} {response.result.status_text}
                      </Badge>
                      <span className="text-sm text-slate-500">{response.result.response_time_ms} ms</span>
                      {response.result.error && (
                        <span className="text-sm text-red-600">{response.result.error}</span>
                      )}
                    </div>
                    <div className="flex gap-1 border-b border-slate-200">
                      {(["body", "headers"] as const).map((tab) => (
                        <button
                          key={tab}
                          onClick={() => setResponseTab(tab)}
                          className={`px-3 py-1.5 text-sm font-medium capitalize ${
                            responseTab === tab
                              ? "border-b-2 border-brand-500 text-brand-700"
                              : "text-slate-500 hover:text-slate-700"
                          }`}
                        >
                          {tab}
                        </button>
                      ))}
                    </div>
                    {responseTab === "body" ? (
                      <pre className="max-h-96 overflow-auto rounded-lg bg-slate-900 p-4 text-sm text-slate-50">
                        {formatResponseBody(response.result.body)}
                      </pre>
                    ) : (
                      <pre className="max-h-96 overflow-auto rounded-lg bg-slate-50 p-4 text-sm text-slate-900">
                        {JSON.stringify(response.result.headers, null, 2)}
                      </pre>
                    )}
                  </div>
                ) : (
                  <EmptyState
                    icon={Send}
                    title="Ready to send"
                    description="Build your request and hit Send to see the response here."
                  />
                )}
              </div>
            </Card>
          ) : (
            <Card className="flex flex-1 items-center justify-center">
              <EmptyState
                icon={Globe}
                title="Select or create a request"
                description={
                  selectedCollectionId
                    ? "Choose a request from the sidebar or create a new one."
                    : "Select a collection from the sidebar to get started."
                }
                action={
                  selectedCollectionId
                    ? { label: "New Request", onClick: newRequest }
                    : undefined
                }
              />
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}

function KeyValueEditor({
  pairs,
  onChange,
  onAdd,
  onRemove,
}: {
  pairs: KeyValuePair[];
  onChange: (index: number, key: "key" | "value" | "enabled", value: string | boolean) => void;
  onAdd: () => void;
  onRemove: (index: number) => void;
}) {
  return (
    <div className="space-y-2">
      <table className="w-full text-sm">
        <thead className="text-left text-slate-500">
          <tr>
            <th className="pb-2 font-medium w-16">Enabled</th>
            <th className="pb-2 font-medium">Key</th>
            <th className="pb-2 font-medium">Value</th>
            <th className="pb-2 font-medium"></th>
          </tr>
        </thead>
        <tbody>
          {pairs.map((pair, i) => (
            <tr key={i} className="border-b border-slate-100">
              <td className="py-2 pr-2">
                <Switch checked={pair.enabled} onCheckedChange={(checked) => onChange(i, "enabled", checked)} />
              </td>
              <td className="py-2 pr-2">
                <Input value={pair.key} onChange={(e) => onChange(i, "key", e.target.value)} className="h-8" />
              </td>
              <td className="py-2 pr-2">
                <Input value={pair.value} onChange={(e) => onChange(i, "value", e.target.value)} className="h-8" />
              </td>
              <td className="py-2">
                <Button variant="ghost" size="sm" onClick={() => onRemove(i)}>
                  <X className="h-4 w-4 text-red-500" />
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <Button variant="secondary" size="sm" onClick={onAdd}>
        <Plus className="mr-1 h-4 w-4" /> Add
      </Button>
    </div>
  );
}

function AuthEditor({
  authType,
  authConfig,
  onChange,
}: {
  authType: AuthType;
  authConfig: AuthConfig;
  onChange: (type: AuthType, config: AuthConfig) => void;
}) {
  const update = (key: string, value: string) => {
    onChange(authType, { ...authConfig, [key]: value } as AuthConfig);
  };

  return (
    <div className="space-y-4">
      <select
        value={authType}
        onChange={(e) => onChange(e.target.value as AuthType, authConfig)}
        className="h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-900 focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
      >
        {AUTH_TYPES.map((t) => (
          <option key={t} value={t}>
            {t === "none" ? "No Auth" : t.replace("_", " ").replace(/\b\w/g, (l) => l.toUpperCase())}
          </option>
        ))}
      </select>

      {authType === "bearer" && (
        <Input
          value={authConfig.bearer_token || ""}
          onChange={(e) => update("bearer_token", e.target.value)}
          placeholder="Bearer token"
        />
      )}
      {authType === "basic" && (
        <div className="grid gap-2 sm:grid-cols-2">
          <Input
            value={authConfig.username || ""}
            onChange={(e) => update("username", e.target.value)}
            placeholder="Username"
          />
          <Input
            value={authConfig.password || ""}
            onChange={(e) => update("password", e.target.value)}
            placeholder="Password"
            type="password"
          />
        </div>
      )}
      {authType === "api_key" && (
        <div className="space-y-2">
          <div className="grid gap-2 sm:grid-cols-3">
            <Input value={authConfig.api_key || ""} onChange={(e) => update("api_key", e.target.value)} placeholder="Key name" />
            <Input value={authConfig.api_value || ""} onChange={(e) => update("api_value", e.target.value)} placeholder="Value" />
            <select
              value={authConfig.api_location || "header"}
              onChange={(e) => update("api_location", e.target.value)}
              className="h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-900 focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
            >
              <option value="header">Header</option>
              <option value="query">Query</option>
            </select>
          </div>
        </div>
      )}
    </div>
  );
}

function BodyEditor({
  bodyType,
  bodyContent,
  onChange,
}: {
  bodyType: BodyType;
  bodyContent: string;
  onChange: (type: BodyType, content: string) => void;
}) {
  return (
    <div className="space-y-3">
      <select
        value={bodyType}
        onChange={(e) => onChange(e.target.value as BodyType, bodyContent)}
        className="h-10 rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-900 focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
      >
        {BODY_TYPES.map((t) => (
          <option key={t} value={t}>
            {t === "none" ? "None" : t.replace("_", " ").replace(/\b\w/g, (l) => l.toUpperCase())}
          </option>
        ))}
      </select>
      {bodyType !== "none" && (
        <textarea
          value={bodyContent}
          onChange={(e) => onChange(bodyType, e.target.value)}
          placeholder="Request body"
          className="h-48 w-full rounded-lg border border-slate-300 p-3 font-mono text-sm text-slate-900 focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
        />
      )}
    </div>
  );
}
