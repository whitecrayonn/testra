"use client";

import { useState, useEffect, useCallback } from "react";
import { Bell, Check, Trash2 } from "lucide-react";
import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import { listNotifications, markRead, deleteNotification } from "@/features/notifications/api";
import type { Notification } from "@/types/notifications";

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [cursor, setCursor] = useState<string | undefined>(undefined);
  const [hasMore, setHasMore] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<"all" | "unread">("all");

  const fetchNotifications = useCallback(
    async (append = false, nextCursor?: string) => {
      if (append) {
        setLoadingMore(true);
      } else {
        setLoading(true);
      }
      setError(null);
      try {
        const read = filter === "unread" ? true : undefined;
        const { data, meta } = await listNotifications(nextCursor, read);
        if (append) {
          setNotifications((prev) => [...prev, ...data]);
        } else {
          setNotifications(data);
        }
        setCursor(meta.next_cursor);
        setHasMore(meta.has_more);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load notifications");
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [filter],
  );

  useEffect(() => {
    fetchNotifications();
  }, [fetchNotifications]);

  async function handleMarkRead(id: string, read: boolean) {
    try {
      await markRead(id, read);
      setNotifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, read } : n)),
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update notification");
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Delete this notification?")) return;
    try {
      await deleteNotification(id);
      setNotifications((prev) => prev.filter((n) => n.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete notification");
    }
  }

  function formatDate(iso: string) {
    return new Date(iso).toLocaleString();
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Notifications"
        description="View and manage your in-app notifications."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Notifications" },
        ]}
      />

      {error && (
        <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>
      )}

      <div className="flex gap-2">
        <Button
          variant={filter === "all" ? "primary" : "secondary"}
          size="sm"
          onClick={() => setFilter("all")}
        >
          All
        </Button>
        <Button
          variant={filter === "unread" ? "primary" : "secondary"}
          size="sm"
          onClick={() => setFilter("unread")}
        >
          Unread
        </Button>
      </div>

      {loading ? (
        <Skeleton className="h-24 w-full" count={3} />
      ) : notifications.length === 0 ? (
        <EmptyState
          icon={Bell}
          title="No notifications"
          description="You're all caught up! New notifications will appear here."
        />
      ) : (
        <div className="space-y-3">
          {notifications.map((n) => (
            <Card
              key={n.id}
              className={n.read ? "opacity-75" : undefined}
            >
              <CardContent className="p-4">
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div className="flex-1 space-y-1">
                    <div className="flex items-center gap-2">
                      <h3 className="font-semibold text-slate-900">{n.title}</h3>
                      {!n.read && <Badge variant="info">New</Badge>}
                    </div>
                    <p className="text-sm text-slate-600">{n.body}</p>
                    {n.link && (
                      <Link
                        href={n.link}
                        className="text-sm font-medium text-brand-600 hover:underline"
                      >
                        View details
                      </Link>
                    )}
                    <p className="text-xs text-slate-400">{formatDate(n.created_at)}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleMarkRead(n.id, !n.read)}
                      aria-label={n.read ? "Mark unread" : "Mark read"}
                    >
                      {n.read ? <Check className="h-4 w-4" /> : <Check className="h-4 w-4" />}
                      {n.read ? "Mark unread" : "Mark read"}
                    </Button>
                    <Button
                      variant="danger"
                      size="sm"
                      onClick={() => handleDelete(n.id)}
                      aria-label="Delete notification"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
          {hasMore && (
            <Button
              variant="secondary"
              className="w-full"
              loading={loadingMore}
              onClick={() => fetchNotifications(true, cursor)}
            >
              Load more
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
