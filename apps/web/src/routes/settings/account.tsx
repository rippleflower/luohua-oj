import { AppShell } from "../../components/layout/app-shell";
import { useMeSettings } from "../../features/auth/hooks";

export function AccountSettingsRoute() {
  const { data } = useMeSettings();

  return (
    <AppShell title="Account Settings" subtitle="settings">
      <div className="grid max-w-2xl gap-4 rounded-3xl border border-slate-200 bg-white/85 p-6">
        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
          <p className="text-sm text-slate-500">Username</p>
          <p className="mt-1 font-semibold text-slate-950">{data?.profile.username ?? "-"}</p>
        </div>
        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
          <p className="text-sm text-slate-500">Email</p>
          <p className="mt-1 font-semibold text-slate-950">{data?.profile.email ?? "-"}</p>
        </div>
        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
          Account identity is managed centrally. This screen is intentionally read-only in the first pass.
        </div>
      </div>
    </AppShell>
  );
}
