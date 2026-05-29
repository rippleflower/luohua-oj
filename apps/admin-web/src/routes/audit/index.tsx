import { useQuery } from "@tanstack/react-query";

import { AdminShell } from "../../components/layout/admin-shell";
import { listAuditEvents } from "../../features/admin/api";

export function AuditRoute() {
  const { data = [] } = useQuery({
    queryKey: ["admin", "audit"],
    queryFn: listAuditEvents,
    retry: false,
  });

  return (
    <AdminShell
      title="审计日志"
      sidebar={
        <section className="rounded-3xl border border-slate-200 bg-white p-5">
          <h2 className="text-lg font-semibold">审计上下文</h2>
          <div className="mt-4 grid gap-2 text-sm text-slate-600">
            <p>事件总数：{data.length}</p>
            <p>最近记录：{data[0] ? new Date(data[0].createdAt).toLocaleString() : "-"}</p>
          </div>
        </section>
      }
    >
      <div className="grid gap-3">
        {data.map((event) => (
          <article key={event.id} className="rounded-3xl border border-slate-200 bg-white p-5">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="font-semibold text-slate-950">{event.action}</p>
                <p className="mt-1 text-sm text-slate-500">
                  {event.actor.username ?? "system"} · {event.targetType}:{event.targetId}
                </p>
              </div>
              <span className="text-xs text-slate-400">{new Date(event.createdAt).toLocaleString()}</span>
            </div>
            {event.reason ? <p className="mt-3 text-sm leading-6 text-slate-600">{event.reason}</p> : null}
          </article>
        ))}
      </div>
    </AdminShell>
  );
}
