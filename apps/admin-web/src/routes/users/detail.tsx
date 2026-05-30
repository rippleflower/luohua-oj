import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { AdminShell } from "../../components/layout/admin-shell";
import { getUser, updateUser, updateUserRole } from "../../features/admin/api";

export function UserDetailRoute({ userId }: { userId: string }) {
  const queryClient = useQueryClient();
  const { data } = useQuery({
    queryKey: ["admin", "users", userId],
    queryFn: () => getUser(userId),
    retry: false,
  });
  const [message, setMessage] = useState("");

  if (!data) {
    return <AdminShell title="用户详情"><p>Loading...</p></AdminShell>;
  }
  const profile = data.profile;
  const status = data.status;

  async function handleProfileSave(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    await updateUser(userId, {
      status: String(formData.get("status") ?? status),
      displayName: String(formData.get("displayName") ?? profile.displayName),
      bio: String(formData.get("bio") ?? profile.bio),
      avatarUrl: String(formData.get("avatarUrl") ?? profile.avatarUrl),
      reason: String(formData.get("reason") ?? "manual update"),
    });
    await queryClient.invalidateQueries({ queryKey: ["admin", "users"] });
    await queryClient.invalidateQueries({ queryKey: ["admin", "users", userId] });
    setMessage("Profile updated.");
  }

  async function handleRoleChange(role: string) {
    await updateUserRole(userId, role, "role updated in admin console");
    await queryClient.invalidateQueries({ queryKey: ["admin", "users"] });
    await queryClient.invalidateQueries({ queryKey: ["admin", "users", userId] });
    setMessage(`Role changed to ${role}.`);
  }

  return (
    <AdminShell title="用户详情">
      <div className="grid gap-4 lg:grid-cols-[minmax(0,1.15fr)_minmax(280px,0.85fr)]">
        <form className="grid gap-4 rounded-3xl border border-slate-200 bg-white p-5" onSubmit={handleProfileSave}>
          <input className="rounded-2xl border border-slate-200 px-4 py-3" defaultValue={data.profile.displayName} name="displayName" />
          <textarea className="min-h-32 rounded-2xl border border-slate-200 px-4 py-3" defaultValue={data.profile.bio} name="bio" />
          <input className="rounded-2xl border border-slate-200 px-4 py-3" defaultValue={data.profile.avatarUrl} name="avatarUrl" />
          <select className="rounded-2xl border border-slate-200 px-4 py-3" defaultValue={data.status} name="status">
            <option value="ACTIVE">ACTIVE</option>
            <option value="SUSPENDED">SUSPENDED</option>
          </select>
          <input className="rounded-2xl border border-slate-200 px-4 py-3" defaultValue="manual update" name="reason" />
          <button className="rounded-full bg-slate-950 px-4 py-3 text-sm font-semibold text-white" type="submit">保存资料</button>
          {message ? <p className="text-sm text-emerald-700">{message}</p> : null}
        </form>
        <aside className="grid gap-4">
          <div className="rounded-3xl border border-slate-200 bg-white p-5">
            <h2 className="text-lg font-semibold">角色</h2>
            <div className="mt-4 flex flex-wrap gap-2">
              {["USER", "ADMIN", "SUPER_ADMIN"].map((role) => (
                <button key={role} className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold" type="button" onClick={() => handleRoleChange(role)}>
                  {role}
                </button>
              ))}
            </div>
          </div>
          <div className="rounded-3xl border border-slate-200 bg-white p-5">
            <h2 className="text-lg font-semibold">统计</h2>
            <div className="mt-4 grid gap-2 text-sm text-slate-600">
              <p>已解决：{data.stats.solvedCount}</p>
              <p>提交数：{data.stats.submissionCount}</p>
              <p>Accepted：{data.stats.acceptedCount}</p>
            </div>
          </div>
        </aside>
      </div>
    </AdminShell>
  );
}
