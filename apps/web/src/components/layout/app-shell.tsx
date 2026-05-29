import type { ReactNode } from "react";
import { webRoutes, webSectionForPath } from "@oj/shared";
import { useAuthUser } from "../../features/auth/hooks";
import { env } from "../../lib/env";
import { useLocale } from "../../lib/locale";

type AppShellProps = {
  title: string;
  subtitle: string;
  action?: ReactNode;
  children: ReactNode;
};

export function AppShell({ title, subtitle, action, children }: AppShellProps) {
  const { locale, setLocale } = useLocale();
  const { data: viewer } = useAuthUser();
  const pathname = typeof window === "undefined" ? "/" : window.location.pathname;
  const activeSection = webSectionForPath(pathname);
  const adminHref = env.adminBaseUrl.replace(/\/$/, "") || "/";
  const isAdminViewer = viewer?.role === "ADMIN" || viewer?.role === "SUPER_ADMIN";
  const navItems = [
    { href: webRoutes.home, section: "home", label: locale === "zh" ? "首页" : "Home" },
    { href: webRoutes.problems, section: "problems", label: locale === "zh" ? "题库" : "Problems" },
    { href: webRoutes.contests, section: "contests", label: locale === "zh" ? "比赛" : "Contests" },
    { href: webRoutes.submissions, section: "submissions", label: locale === "zh" ? "提交" : "Submissions" },
    { href: webRoutes.me, section: "me", label: locale === "zh" ? "我的" : "Me" },
  ];

  return (
    <main className="min-h-screen bg-transparent text-slate-950">
      <section className="border-b border-slate-200/70 bg-white/70 backdrop-blur-xl">
        <div className="mx-auto flex max-w-7xl items-center justify-between gap-4 px-6 py-3.5">
          <div className="flex items-center gap-6">
            <a className="flex items-center gap-2.5" href={webRoutes.home}>
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-slate-900/10 bg-slate-950 font-mono text-sm text-white">
                lu
              </span>
              <div>
                <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">{locale === "zh" ? "竞赛判题系统" : "Online Judge"}</p>
                <p className="text-sm font-semibold text-slate-900">luooj</p>
              </div>
            </a>
            <nav className="hidden items-center gap-1.5 md:flex">
              {navItems.map((item) => {
                const active = activeSection === item.section;
                return (
                  <a
                    key={item.href}
                    className={`rounded-full px-3.5 py-1.5 text-sm transition ${
                      active ? "bg-slate-950 text-white" : "text-slate-600 hover:bg-white hover:text-slate-950"
                    }`}
                    href={item.href}
                  >
                    {item.label}
                  </a>
                );
              })}
            </nav>
          </div>
          <div className="flex items-center gap-2.5">
            {isAdminViewer ? (
              <a
                className="rounded-full border border-amber-200 bg-amber-50 px-3 py-1.5 text-sm font-semibold text-amber-800 hover:bg-amber-100"
                href={adminHref}
              >
                {locale === "zh" ? "管理端" : "Admin Console"}
              </a>
            ) : null}
            <a
              className="rounded-full border border-slate-200 bg-white/80 px-3 py-1.5 text-sm font-semibold text-slate-700 hover:text-slate-950"
              href={viewer ? webRoutes.me : webRoutes.login}
            >
              {viewer ? viewer.displayName : locale === "zh" ? "登录" : "Log In"}
            </a>
            <button
              className="rounded-full border border-slate-200 bg-white/80 px-3 py-1.5 font-mono text-[11px] text-slate-600 hover:text-slate-950"
              onClick={() => setLocale(locale === "zh" ? "en" : "zh")}
              type="button"
            >
              {locale === "zh" ? "中文 / EN" : "EN / 中文"}
            </button>
            <div className="hidden rounded-full border border-slate-200 bg-white/80 px-3 py-1.5 font-mono text-[11px] text-slate-500 lg:block">
              {pathname}
            </div>
          </div>
        </div>
      </section>
      <section className="mx-auto max-w-7xl px-6 py-6">
        <div className="mb-6 flex flex-col gap-3 border-b border-slate-200/80 pb-5 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p className="font-mono text-[11px] uppercase tracking-[0.3em] text-slate-500">{subtitle}</p>
            <h1 className="mt-2.5 font-['Newsreader'] text-[2.25rem] leading-none text-slate-950 sm:text-[3.25rem]">{title}</h1>
          </div>
          {action}
        </div>
        {children}
      </section>
    </main>
  );
}
