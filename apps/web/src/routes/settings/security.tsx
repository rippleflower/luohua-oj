import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";

import { AppShell } from "../../components/layout/app-shell";
import { changePassword, revokeSession } from "../../features/auth/api";
import { useMeSettings, useSessions } from "../../features/auth/hooks";

export function SecuritySettingsRoute() {
  const queryClient = useQueryClient();
  const { data: settings } = useMeSettings();
  const { data: sessions = [] } = useSessions();
  const [message, setMessage] = useState("");

  async function handlePasswordSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    await changePassword({
      currentPassword: String(formData.get("currentPassword") ?? ""),
      newPassword: String(formData.get("newPassword") ?? ""),
      confirmPassword: String(formData.get("confirmPassword") ?? ""),
    });
    setMessage("Password updated. Other sessions were revoked.");
    await queryClient.invalidateQueries({ queryKey: ["auth"] });
  }

  async function handleRevoke(sessionId: string) {
    await revokeSession(sessionId);
    await queryClient.invalidateQueries({ queryKey: ["auth", "sessions"] });
    await queryClient.invalidateQueries({ queryKey: ["me"] });
  }

  return (
    <AppShell title="Security Settings" subtitle="settings">
      <div className="grid gap-4 lg:grid-cols-2">
        <form className="grid gap-4 rounded-3xl border border-slate-200 bg-white/85 p-6" onSubmit={handlePasswordSubmit}>
          <input className="rounded-2xl border border-slate-200 px-4 py-3" name="currentPassword" placeholder="Current password" type="password" />
          <input className="rounded-2xl border border-slate-200 px-4 py-3" name="newPassword" placeholder="New password" type="password" />
          <input className="rounded-2xl border border-slate-200 px-4 py-3" name="confirmPassword" placeholder="Confirm password" type="password" />
          <button className="rounded-full bg-slate-950 px-4 py-3 text-sm font-semibold text-white" type="submit">Change password</button>
          {message ? <p className="text-sm text-emerald-700">{message}</p> : null}
        </form>
        <div className="grid gap-3 rounded-3xl border border-slate-200 bg-white/85 p-6">
          <h2 className="text-lg font-semibold text-slate-950">Sessions</h2>
          {settings ? <p className="text-sm text-slate-500">{settings.user.displayName}</p> : null}
          {sessions.map((session) => (
            <div key={session.id} className="flex items-center justify-between rounded-2xl border border-slate-200 p-4">
              <div>
                <p className="font-semibold text-slate-950">{session.current ? "Current device" : "Active device"}</p>
                <p className="mt-1 text-sm text-slate-500">{session.ip}</p>
              </div>
              {!session.current ? (
                <button className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold" type="button" onClick={() => handleRevoke(session.id)}>
                  Revoke
                </button>
              ) : null}
            </div>
          ))}
        </div>
      </div>
    </AppShell>
  );
}
