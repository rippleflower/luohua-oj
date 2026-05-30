import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";

import { AppShell } from "../../components/layout/app-shell";
import { updateProfile } from "../../features/auth/api";
import { useMeSettings } from "../../features/auth/hooks";

export function ProfileSettingsRoute() {
  const queryClient = useQueryClient();
  const { data } = useMeSettings();
  const [message, setMessage] = useState("");

  if (!data) {
    return <AppShell title="Profile Settings" subtitle="settings"><p>Sign in required.</p></AppShell>;
  }

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    await updateProfile({
      displayName: String(formData.get("displayName") ?? ""),
      bio: String(formData.get("bio") ?? ""),
      avatarUrl: String(formData.get("avatarUrl") ?? ""),
    });
    await queryClient.invalidateQueries({ queryKey: ["me"] });
    setMessage("Saved.");
  }

  return (
    <AppShell title="Profile Settings" subtitle="settings">
      <form className="grid max-w-2xl gap-4 rounded-3xl border border-slate-200 bg-white/85 p-6" onSubmit={handleSubmit}>
        <input className="rounded-2xl border border-slate-200 px-4 py-3" name="displayName" defaultValue={data.profile.displayName} />
        <textarea className="min-h-32 rounded-2xl border border-slate-200 px-4 py-3" name="bio" defaultValue={data.profile.bio} />
        <input className="rounded-2xl border border-slate-200 px-4 py-3" name="avatarUrl" defaultValue={data.profile.avatarUrl} />
        <button className="rounded-full bg-slate-950 px-4 py-3 text-sm font-semibold text-white" type="submit">Save profile</button>
        {message ? <p className="text-sm text-emerald-700">{message}</p> : null}
      </form>
    </AppShell>
  );
}
