import { useEffect, type ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";

import { permissionKeys } from "@oj/shared";
import { getAuthMe } from "../../features/auth/api";
import { env } from "../../lib/env";

type AdminShellProps = {
  title: string;
  children: ReactNode;
};

const navItems = [
  { href: "/", label: "总览", permission: "dashboard.view" },
  { href: "/users", label: "用户与权限", permission: "users.view" },
  { href: "/problems", label: "题库管理", permission: "problems.view" },
  { href: "/contests", label: "比赛管理", permission: "contests.view" },
  { href: "/submissions", label: "提交与判题", permission: "submissions.view" },
  { href: "/announcements", label: "公告与运营", permission: "announcements.view" },
  { href: "/system/settings", label: "系统配置", permission: "system.view" },
  { href: "/audit", label: "审计日志", permission: "audit.view" },
] as const;

export function AdminShell({ title, children }: AdminShellProps) {
  const pathname = window.location.pathname.replace(/\/$/, "") || "/";
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
    <main className="min-h-screen text-slate-950">
      <div className="grid min-h-screen w-full gap-4 px-3 py-4 sm:px-4 lg:grid-cols-[260px_minmax(0,1fr)]">
        <aside className="rounded-[2rem] border border-orange-950/10 bg-white/80 p-5 shadow-[0_20px_80px_rgba(120,53,15,0.08)]">
          <a className="block rounded-2xl bg-slate-950 px-4 py-4 text-white" href="/">
            <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-orange-200">luooj admin</p>
            <p className="mt-2 text-xl font-semibold">{viewer?.displayName ?? "后台管理"}</p>
          </a>
          <nav className="mt-5 grid gap-1.5">
            {navItems
              .filter((item) => viewer?.role === "SUPER_ADMIN" || allowed.has(item.permission))
              .map((item) => {
                const active = item.href === "/" ? pathname === "/" : pathname.startsWith(item.href);
                return (
                  <a key={item.href} className={`rounded-2xl px-4 py-3 text-sm font-semibold ${active ? "bg-orange-600 text-white" : "text-slate-700 hover:bg-orange-50"}`} href={item.href}>
                    {item.label}
                  </a>
                );
              })}
          </nav>
        </aside>
        <section className="rounded-[2rem] border border-white/60 bg-white/78 p-6 shadow-[0_20px_80px_rgba(120,53,15,0.08)]">
          <div className="mb-5 flex items-end justify-between gap-4 border-b border-slate-200 pb-4">
            <div>
              <p className="font-mono text-[11px] uppercase tracking-[0.3em] text-slate-500">operations console</p>
              <h1 className="mt-2 font-['Newsreader'] text-5xl leading-none">{title}</h1>
            </div>
            <div className="flex items-center gap-2">
              <a
                className="rounded-full border border-orange-200 bg-orange-50 px-4 py-2 text-sm font-semibold text-orange-700 transition hover:bg-orange-100"
                href={previewHref}
              >
                前台
              </a>
              {viewer ? <div className="rounded-full border border-slate-200 px-4 py-2 text-sm text-slate-600">{viewer.role}</div> : null}
            </div>
          </div>
          {children}
        </section>
      </div>
    </main>
  );
}

function previewPathFor(pathname: string): string {
  if (pathname.startsWith("/problems")) {
    return "/problems";
  }
  if (pathname.startsWith("/contests")) {
    return "/contests";
  }
  if (pathname.startsWith("/submissions") || pathname.startsWith("/judge/queue")) {
    return "/submissions";
  }
  return "/";
}
