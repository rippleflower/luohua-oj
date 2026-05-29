import { useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { AdminShell } from "../../components/layout/admin-shell";
import { getSystemSettings, updateSystemSettings } from "../../features/admin/api";

export function SystemSettingsRoute({ section }: { section: "settings" | "storage" | "judge" }) {
  const queryClient = useQueryClient();
  const { data } = useQuery({
    queryKey: ["admin", "system", "settings"],
    queryFn: getSystemSettings,
    retry: false,
  });
  const [message, setMessage] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    await updateSystemSettings({
      registrationEnabled: String(formData.get("registrationEnabled")) === "true",
      judgeQueuePaused: String(formData.get("judgeQueuePaused")) === "true",
      storageMode: String(formData.get("storageMode")) as "LOCAL" | "S3",
    });
    await queryClient.invalidateQueries({ queryKey: ["admin", "system", "settings"] });
    setMessage("System settings updated.");
  }

  return (
    <AdminShell
      title={section === "storage" ? "存储配置" : section === "judge" ? "判题配置" : "系统配置"}
      sidebar={
        <>
          <div className="rounded-3xl border border-slate-200 bg-white p-5">
            <h2 className="text-lg font-semibold">运行时摘要</h2>
            <div className="mt-4 grid gap-2 text-sm text-slate-600">
              <p>Source Root: {data?.sourceRoot ?? "-"}</p>
              <p>Redis: {data?.redisAddr ?? "-"}</p>
              <p>Updated: {data ? new Date(data.updatedAt).toLocaleString() : "-"}</p>
            </div>
          </div>
        </>
      }
    >
      <form className="grid gap-4 rounded-3xl border border-slate-200 bg-white p-5" onSubmit={handleSubmit}>
        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
          独立后台把系统开关、对象存储和 judge 队列放在同一组摘要配置页里，避免分散到隐藏入口。
        </div>
        <select className="rounded-2xl border border-slate-200 px-4 py-3" defaultValue={String(data?.registrationEnabled ?? true)} name="registrationEnabled">
          <option value="true">允许注册</option>
          <option value="false">关闭注册</option>
        </select>
        <select className="rounded-2xl border border-slate-200 px-4 py-3" defaultValue={String(data?.judgeQueuePaused ?? false)} name="judgeQueuePaused">
          <option value="false">判题队列运行中</option>
          <option value="true">暂停判题队列</option>
        </select>
        <select className="rounded-2xl border border-slate-200 px-4 py-3" defaultValue={data?.storageMode ?? "LOCAL"} name="storageMode">
          <option value="LOCAL">LOCAL</option>
          <option value="S3">S3</option>
        </select>
        <button className="rounded-full bg-slate-950 px-4 py-3 text-sm font-semibold text-white" type="submit">保存配置</button>
        {message ? <p className="text-sm text-emerald-700">{message}</p> : null}
      </form>
    </AdminShell>
  );
}
