import { SettingsNav } from "@/components/dashboard/settings-nav";

export default function SettingsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">Settings</h1>
        <p className="mt-1 text-sm text-slate-500">Manage your account, workspace, and organization preferences.</p>
      </div>
      <SettingsNav />
      <div className="min-h-[400px]">{children}</div>
    </div>
  );
}
