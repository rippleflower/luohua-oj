import { useEffect, type ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { adminRoutes, isRouteActive, normalizePathname, webRoutes } from "@oj/shared";

import { permissionKeys } from "@oj/shared";
import { getAuthMe } from "../../features/auth/api";
import { env } from "../../lib/env";

type AdminShellProps = {
  title: string;
  children: ReactNode;
  sidebar?: ReactNode;
};

const navItems = [
  { href: adminRoutes.dashboard, label: "总览", permission: "dashboard.view" },
  { href: adminRoutes.users, label: "用户与权限", permission: "users.view" },
  { href: adminRoutes.problems, label: "题库管理", permission: "problems.view" },
  { href: adminRoutes.contests, label: "比赛管理", permission: "contests.view" },
  { href: adminRoutes.submissions, label: "提交与判题", permission: "submissions.view" },
  { href: adminRoutes.announcements, label: "公告与运营", permission: "announcements.view" },
  { href: adminRoutes.systemSettings, label: "系统配置", permission: "system.view" },
  { href: adminRoutes.audit, label: "审计日志", permission: "audit.view" },
] as const;

export function AdminShell({ title, children, sidebar }: AdminShellProps) {
  const pathname = normalizePathname(window.location.pathname);
  const webBaseUrl = env.webBaseUrl.replace(/\/$/, "");
  const { data: viewer, error, isPending, isError } = useQuery({
    queryKey: ["admin", "viewer"],
    queryFn: getAuthMe,
    retry: false,
  });
  const unauthorized =
    error instanceof Error &&
    (/\b401\b/.test(error.message) ||
      /authentication required|unauthorized|not authenticated/i.test(error.message));

  useEffect(() => {
    if (unauthorized) {
      window.location.replace("/login");
    }
  }, [unauthorized]);

  if (isPending) {
    return (
      <main className="flex min-h-screen items-center justify-center px-6 text-slate-600">
        正在加载后台身份信息...
      </main>
    );
  }

  if (isError) {
    if (unauthorized) {
      return null;
    }
    const errorMessage = error instanceof Error ? error.message : "unknown error";
    return (
      <main className="flex min-h-screen items-center justify-center px-6">
        <div className="max-w-xl rounded-2xl border border-rose-200 bg-rose-50 px-5 py-4 text-sm text-rose-700">
          后台身份校验失败：{errorMessage}。请确认 API 服务已启动，或检查 VITE_API_BASE_URL 配置。
        </div>
      </main>
    );
  }

  if (!viewer) {
    return null;
  }

  const allowed = new Set(viewer.role === "SUPER_ADMIN" ? permissionKeys : viewer.permissions ?? []);
  const previewPath = previewPathFor(pathname);
  const previewHref = webBaseUrl === "" ? previewPath : `${webBaseUrl}${previewPath}`;

  return (
    <main className="min-h-screen bg-[linear-gradient(180deg,#fff7ed_0%,#fff 28%,#f8fafc_100%)] text-slate-950">
      <header className="border-b border-orange-950/10 bg-white/85 backdrop-blur-xl">
        <div className="mx-auto flex max-w-[1600px] flex-col gap-4 px-4 py-4 lg:px-6">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <a className="block rounded-3xl bg-slate-950 px-5 py-4 text-white shadow-[0_20px_80px_rgba(15,23,42,0.12)]" href={adminRoutes.dashboard}>
              <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-orange-200">luooj admin</p>
              <div className="mt-2 flex flex-wrap items-end gap-3">
                <p className="text-2xl font-semibold">{viewer?.displayName ?? "后台管理"}</p>
                <span className="rounded-full border border-white/15 px-3 py-1 text-xs text-orange-100">{viewer.role}</span>
              </div>
            </a>
            <div className="flex flex-wrap items-center gap-2">
              <a
                className="rounded-full border border-orange-200 bg-orange-50 px-4 py-2 text-sm font-semibold text-orange-700 transition hover:bg-orange-100"
                href={previewHref}
              >
                前台
              </a>
              <div className="rounded-full border border-slate-200 bg-white/90 px-4 py-2 text-sm text-slate-600">{title}</div>
            </div>
          </div>
          <nav className="flex flex-wrap gap-2">
            {navItems
              .filter((item) => viewer?.role === "SUPER_ADMIN" || allowed.has(item.permission))
              .map((item) => {
                const active = isRouteActive(pathname, item.href);
                return (
                  <a key={item.href} className={`rounded-full px-4 py-2.5 text-sm font-semibold ${active ? "bg-orange-600 text-white" : "bg-white/80 text-slate-700 hover:bg-orange-50"}`} href={item.href}>
                    {item.label}
                  </a>
                );
              })}
          </nav>
        </div>
      </header>
      <section className="mx-auto max-w-[1600px] px-4 py-5 lg:px-6">
        <div className={`grid gap-4 ${sidebar ? "xl:grid-cols-[minmax(0,1fr)_340px]" : ""}`}>
          <section className="rounded-[2rem] border border-white/60 bg-white/84 p-6 shadow-[0_20px_80px_rgba(120,53,15,0.08)]">
            <div className="mb-5 flex items-end justify-between gap-4 border-b border-slate-200 pb-4">
            <div>
              <p className="font-mono text-[11px] uppercase tracking-[0.3em] text-slate-500">operations console</p>
              <h1 className="mt-2 font-['Newsreader'] text-5xl leading-none">{title}</h1>
            </div>
          </div>
            {children}
          </section>
          {sidebar ? (
            <aside className="grid gap-4 self-start">
              {sidebar}
            </aside>
          ) : null}
        </div>
      </section>
    </main>
  );
}

function previewPathFor(pathname: string): string {
  if (pathname.startsWith("/problems")) {
    return webRoutes.problems;
  }
  if (pathname.startsWith("/contests")) {
    return webRoutes.contests;
  }
  if (pathname.startsWith("/submissions") || pathname.startsWith("/judge/queue")) {
    return webRoutes.submissions;
  }
  return webRoutes.home;
}
