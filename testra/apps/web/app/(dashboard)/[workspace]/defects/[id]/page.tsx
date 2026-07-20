"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Save, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { LinkButton } from "@/components/ui/link-button";
import { PageHeader } from "@/components/ui/page-header";
import { getDefect, updateDefect, deleteDefect } from "@/features/defects/api";
import type { Defect } from "@/types/defects";

const statusVariants: Record<string, "neutral" | "info" | "success" | "danger" | "warning"> = {
  open: "info",
  in_progress: "warning",
  resolved: "success",
  closed: "neutral",
  rejected: "danger",
};

const severityVariants: Record<string, "default" | "info" | "warning" | "danger"> = {
  low: "default",
  medium: "info",
  high: "warning",
  critical: "danger",
};

const priorityVariants: Record<string, "default" | "info" | "warning" | "danger"> = {
  low: "default",
  medium: "info",
  high: "warning",
  critical: "danger",
};

export default function DefectDetailPage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const [defect, setDefect] = useState<Defect | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [status, setStatus] = useState<Defect["status"]>();
  const [severity, setSeverity] = useState<Defect["severity"]>();
  const [priority, setPriority] = useState<Defect["priority"]>();

  useEffect(() => {
    async function load() {
      try {
        const d = await getDefect(id);
        setDefect(d);
        setTitle(d.title);
        setDescription(d.description);
        setStatus(d.status);
        setSeverity(d.severity);
        setPriority(d.priority);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load defect");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [id]);

  async function handleSave() {
    if (!status || !severity || !priority) return;
    setSaving(true);
    setError(null);
    try {
      const updated = await updateDefect(id, {
        title: title.trim(),
        description: description.trim(),
        status,
        severity,
        priority,
      });
      setDefect(updated);
      setTitle(updated.title);
      setDescription(updated.description);
      setStatus(updated.status);
      setSeverity(updated.severity);
      setPriority(updated.priority);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save defect");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    if (!confirm("Are you sure you want to delete this defect?")) return;
    try {
      await deleteDefect(id);
      router.push("/dashboard/defects");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete defect");
    }
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <PageHeader title="Defect" description="Loading defect details..." />
        <Card className="p-8 text-center text-slate-500">Loading...</Card>
      </div>
    );
  }

  if (!defect) {
    return (
      <div className="space-y-4">
        <LinkButton href="/dashboard/defects" variant="ghost" size="sm">
          <ArrowLeft className="mr-2 h-4 w-4" aria-hidden="true" />
          Back to Defects
        </LinkButton>
        <div role="alert">
          <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">
            {error || "Defect not found."}
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={defect.title}
        description="View and manage the defect lifecycle."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Defects", href: "/dashboard/defects" },
          { label: defect.title },
        ]}
        actions={
          <div className="flex flex-wrap gap-2">
            <Button variant="danger" size="sm" onClick={handleDelete}>
              <Trash2 className="mr-2 h-4 w-4" aria-hidden="true" />
              Delete
            </Button>
            <Button onClick={handleSave} loading={saving} size="sm">
              <Save className="mr-2 h-4 w-4" aria-hidden="true" />
              Save
            </Button>
          </div>
        }
      />

      <div className="flex flex-wrap gap-2">
        <Badge variant={statusVariants[defect.status] || "neutral"}>
          {defect.status.replace("_", " ")}
        </Badge>
        <Badge variant={severityVariants[defect.severity] || "default"}>
          {defect.severity}
        </Badge>
        <Badge variant={priorityVariants[defect.priority] || "default"}>
          {defect.priority}
        </Badge>
      </div>

      {error && (
        <div role="alert">
          <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">
            {error}
          </Card>
        </div>
      )}

      <Card className="space-y-4 p-6">
        <div>
          <label htmlFor="defect-title" className="mb-1 block text-sm font-medium text-slate-700">
            Title
          </label>
          <Input
            id="defect-title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div>
            <label htmlFor="defect-status" className="mb-1 block text-sm font-medium text-slate-700">
              Status
            </label>
            <select
              id="defect-status"
              value={status}
              onChange={(e) => setStatus(e.target.value as Defect["status"])}
              className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
            >
              <option value="open">Open</option>
              <option value="in_progress">In Progress</option>
              <option value="resolved">Resolved</option>
              <option value="closed">Closed</option>
              <option value="rejected">Rejected</option>
            </select>
          </div>
          <div>
            <label htmlFor="defect-severity" className="mb-1 block text-sm font-medium text-slate-700">
              Severity
            </label>
            <select
              id="defect-severity"
              value={severity}
              onChange={(e) => setSeverity(e.target.value as Defect["severity"])}
              className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
            >
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </select>
          </div>
          <div>
            <label htmlFor="defect-priority" className="mb-1 block text-sm font-medium text-slate-700">
              Priority
            </label>
            <select
              id="defect-priority"
              value={priority}
              onChange={(e) => setPriority(e.target.value as Defect["priority"])}
              className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
            >
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </select>
          </div>
        </div>

        <div>
          <label htmlFor="defect-description" className="mb-1 block text-sm font-medium text-slate-700">
            Description
          </label>
          <textarea
            id="defect-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={5}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500"
          />
        </div>
      </Card>
    </div>
  );
}
