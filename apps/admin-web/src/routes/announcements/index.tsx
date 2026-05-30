import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { AdminShell } from "../../components/layout/admin-shell";
import { createAnnouncement, listAnnouncements } from "../../features/admin/api";

export function AnnouncementsRoute() {
  const queryClient = useQueryClient();
  const { data = [] } = useQuery({
    queryKey: ["admin", "announcements"],
    queryFn: listAnnouncements,
    retry: false,
  });
  const [message, setMessage] = useState("");

  async function handleCreate(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    await createAnnouncement({
      title: String(formData.get("title") ?? ""),
      content: String(formData.get("content") ?? ""),
      status: String(formData.get("status") ?? "DRAFT") as "DRAFT" | "PUBLISHED",
      audience: String(formData.get("audience") ?? "ALL") as "ALL" | "USERS" | "ADMINS",
    });
    await queryClient.invalidateQueries({ queryKey: ["admin", "announcements"] });
    setMessage("Announcement created.");
  }

  return (
    <AdminShell title="公告与运营">
      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px]">
        <div className="grid gap-3">
          {data.map((announcement) => (
            <article key={announcement.id} className="rounded-3xl border border-slate-200 bg-white p-5">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <h2 className="text-lg font-semibold text-slate-950">{announcement.title}</h2>
                  <p className="mt-2 text-sm leading-6 text-slate-600">{announcement.content}</p>
                </div>
                <span className="rounded-full border border-slate-200 px-3 py-1.5 text-xs font-semibold">{announcement.status}</span>
              </div>
            </article>
          ))}
        </div>
        <form className="grid h-fit gap-4 rounded-3xl border border-slate-200 bg-white p-5" onSubmit={handleCreate}>
          <h2 className="text-lg font-semibold">新建公告</h2>
          <input className="rounded-2xl border border-slate-200 px-4 py-3" name="title" placeholder="标题" />
          <textarea className="min-h-40 rounded-2xl border border-slate-200 px-4 py-3" name="content" placeholder="内容" />
          <select className="rounded-2xl border border-slate-200 px-4 py-3" defaultValue="DRAFT" name="status">
            <option value="DRAFT">DRAFT</option>
            <option value="PUBLISHED">PUBLISHED</option>
          </select>
          <select className="rounded-2xl border border-slate-200 px-4 py-3" defaultValue="ALL" name="audience">
            <option value="ALL">ALL</option>
            <option value="USERS">USERS</option>
            <option value="ADMINS">ADMINS</option>
          </select>
          <button className="rounded-full bg-slate-950 px-4 py-3 text-sm font-semibold text-white" type="submit">创建公告</button>
          {message ? <p className="text-sm text-emerald-700">{message}</p> : null}
        </form>
      </div>
    </AdminShell>
  );
}
