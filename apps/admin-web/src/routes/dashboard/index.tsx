import { useQuery } from "@tanstack/react-query";

import { AdminShell } from "../../components/layout/admin-shell";
import { getAdminDashboard } from "../../features/auth/api";

export function DashboardRoute() {
  const { data } = useQuery({
    queryKey: ["admin", "dashboard"],
    queryFn: getAdminDashboard,
    retry: false,
  });

  return (
    <AdminShell
      title="总览"
      sidebar={
        <>
          <section className="rounded-3xl border border-slate-200 bg-white p-5">
            <h2 className="text-lg font-semibold">关键指标</h2>
            <div className="mt-4 grid gap-3">
              {[
                ["用户", data?.metrics.totalUsers ?? 0],
                ["公开题目", data?.metrics.activeProblems ?? 0],
                ["进行中比赛", data?.metrics.runningContests ?? 0],
                ["24h 提交", data?.metrics.submissions24h ?? 0],
                ["待判队列", data?.metrics.pendingSubmissions ?? 0],
              ].map(([label, value]) => (
                <div key={String(label)} className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                  <p className="text-sm text-slate-500">{label}</p>
                  <p className="mt-2 text-3xl font-semibold text-slate-950">{value}</p>
                </div>
              ))}
            </div>
          </section>
          <section className="rounded-3xl border border-slate-200 bg-white p-5">
            <h2 className="text-lg font-semibold">管理域</h2>
            <div className="mt-4 grid gap-2 text-sm text-slate-600">
              <a className="rounded-2xl border border-slate-200 p-3" href="/users">用户与权限</a>
              <a className="rounded-2xl border border-slate-200 p-3" href="/problems">题库管理</a>
              <a className="rounded-2xl border border-slate-200 p-3" href="/contests">比赛管理</a>
              <a className="rounded-2xl border border-slate-200 p-3" href="/system/settings">系统配置</a>
            </div>
          </section>
        </>
      }
    >
      <section className="rounded-3xl border border-slate-200 bg-white p-5">
        <h2 className="text-lg font-semibold">管理员待处理项</h2>
        <div className="mt-4 grid gap-3">
          {(data?.alerts ?? []).map((alert) => (
            <article key={alert.title} className="rounded-2xl border border-slate-200 p-4">
              <p className="font-semibold text-slate-950">{alert.title}</p>
              <p className="mt-1 text-sm text-slate-500">{alert.description}</p>
            </article>
          ))}
        </div>
      </section>
    </AdminShell>
  );
}
