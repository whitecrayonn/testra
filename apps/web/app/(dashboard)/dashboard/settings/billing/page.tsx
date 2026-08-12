"use client";

import { useEffect, useState } from "react";
import { CreditCard, Loader2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { getSubscription, updateSubscription, listInvoices } from "@/features/billing/api";
import type { Subscription, Invoice } from "@/types/billing";

export default function BillingSettingsPage() {
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [invoices, setInvoices] = useState<Invoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [plan, setPlan] = useState("");
  const [seats, setSeats] = useState(1);
  const [cancel, setCancel] = useState(false);

  useEffect(() => {
    const workspaceId = localStorage.getItem("testra_workspace_id") || "";
    if (!workspaceId) {
      setLoading(false);
      setError("No workspace selected");
      return;
    }
    Promise.all([getSubscription(workspaceId), listInvoices(workspaceId)])
      .then(([sub, invs]) => {
        setSubscription(sub);
        setInvoices(invs);
        setPlan(sub.plan);
        setSeats(sub.seats);
        setCancel(sub.cancel_at_period_end);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load billing"))
      .finally(() => setLoading(false));
  }, []);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    const workspaceId = localStorage.getItem("testra_workspace_id") || "";
    if (!workspaceId || !subscription) return;
    setSaving(true);
    setError(null);
    try {
      const updated = await updateSubscription({
        workspace_id: workspaceId,
        plan,
        seats,
        cancel_at_period_end: cancel,
      });
      setSubscription(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update subscription");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Billing"
        description="Manage your subscription, payment method, and invoice history."
        breadcrumbs={[
          { label: "Dashboard", href: "/dashboard" },
          { label: "Settings", href: "/dashboard/settings" },
          { label: "Billing" },
        ]}
      />

      {error && <Card className="border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</Card>}

      {loading ? (
        <Card className="flex h-40 items-center justify-center">
          <Loader2 className="h-6 w-6 animate-spin text-slate-400" />
        </Card>
      ) : (
        <>
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <CreditCard className="h-5 w-5 text-brand-600" />
                Current plan
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {subscription ? (
                <div className="flex flex-wrap items-center gap-4">
                  <div>
                    <p className="text-sm text-slate-500">Plan</p>
                    <Badge variant="success" className="mt-1 capitalize">
                      {subscription.plan}
                    </Badge>
                  </div>
                  <div>
                    <p className="text-sm text-slate-500">Status</p>
                    <p className="font-medium capitalize">{subscription.status}</p>
                  </div>
                  <div>
                    <p className="text-sm text-slate-500">Seats</p>
                    <p className="font-medium">{subscription.seats}</p>
                  </div>
                </div>
              ) : (
                <EmptyState icon={CreditCard} title="No subscription" description="Set up your subscription below." />
              )}

              <form onSubmit={handleSave} className="space-y-4 rounded-lg border border-slate-200 p-4">
                <div className="grid gap-4 sm:grid-cols-2">
                  <Input label="Plan" value={plan} onChange={(e) => setPlan(e.target.value)} required />
                  <Input
                    label="Seats"
                    type="number"
                    min={1}
                    value={seats}
                    onChange={(e) => setSeats(Number(e.target.value))}
                    required
                  />
                </div>
                <label className="flex items-center gap-2 text-sm text-slate-700">
                  <input type="checkbox" checked={cancel} onChange={(e) => setCancel(e.target.checked)} />
                  Cancel at period end
                </label>
                <Button type="submit" loading={saving}>
                  Save subscription
                </Button>
              </form>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Invoice history</CardTitle>
            </CardHeader>
            <CardContent>
              {invoices.length === 0 ? (
                <EmptyState icon={CreditCard} title="No invoices" description="Invoice history will appear here." />
              ) : (
                <div className="space-y-3">
                  {invoices.map((invoice) => (
                    <div key={invoice.id} className="flex items-center justify-between rounded-lg border border-slate-200 p-4">
                      <div>
                        <p className="font-medium text-slate-900">{invoice.provider_invoice_id || invoice.id.slice(0, 8)}</p>
                        <p className="text-sm text-slate-500">{invoice.period_start} - {invoice.period_end}</p>
                      </div>
                      <div className="text-right">
                        <p className="font-medium">{(invoice.amount_cents / 100).toFixed(2)} {invoice.currency}</p>
                        <Badge variant="neutral" className="mt-1 capitalize">{invoice.status}</Badge>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
