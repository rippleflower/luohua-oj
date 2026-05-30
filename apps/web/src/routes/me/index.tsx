import { AppShell } from "../../components/layout/app-shell";
import { useMeSummary } from "../../features/auth/hooks";
import { useLocale } from "../../lib/locale";

export function MeRoute() {
  const { locale } = useLocale();
  const { data, error, isLoading } = useMeSummary();

  if (isLoading) {
    return <AppShell title={locale === "zh" ? "个人中心" : "Me"} subtitle="account"><p>Loading...</p></AppShell>;
  }

  if (error || !data) {
    return (
      <AppShell title={locale === "zh" ? "个人中心" : "Me"} subtitle="account">
        <div className="rounded-3xl border border-dashed border-slate-300 bg-white/70 p-6 text-sm text-slate-600">
          <p>{locale === "zh" ? "你还没有登录。" : "You are not signed in."}</p>
          <a className="mt-3 inline-flex rounded-full bg-slate-950 px-4 py-2 text-white" href="/login">
            {locale === "zh" ? "去登录" : "Sign in"}
          </a>
        </div>
      </AppShell>
    );
  }

  return (
    <AppShell title={data.user.displayName} subtitle={locale === "zh" ? "个人中心" : "Personal Hub"}>
      <div className="grid gap-4 lg:grid-cols-[minmax(0,1.2fr)_minmax(280px,0.8fr)]">
        <section className="grid gap-4">
          <div className="grid gap-3 rounded-3xl border border-slate-200 bg-white/85 p-5 md:grid-cols-3">
            {[
              [locale === "zh" ? "已解决题目" : "Solved", data.stats.solvedCount],
              [locale === "zh" ? "总提交" : "Submissions", data.stats.submissionCount],
              [locale === "zh" ? "Accepted" : "Accepted", data.stats.acceptedCount],
            ].map(([label, value]) => (
              <div key={String(label)} className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                <p className="text-sm text-slate-500">{label}</p>
                <p className="mt-2 text-3xl font-semibold text-slate-950">{value}</p>
              </div>
            ))}
          </div>
          <div className="rounded-3xl border border-slate-200 bg-white/85 p-5">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold">{locale === "zh" ? "最近提交" : "Recent submissions"}</h2>
              <a className="text-sm font-medium text-slate-600" href="/submissions">{locale === "zh" ? "查看全部" : "View all"}</a>
            </div>
            <div className="mt-4 grid gap-3">
              {data.recentSubmissions.map((item) => (
                <a key={item.id} className="rounded-2xl border border-slate-200 p-4" href={`/submissions/${item.id}`}>
                  <p className="font-semibold text-slate-950">{item.problem.title}</p>
                  <p className="mt-1 text-sm text-slate-500">{item.status}</p>
                </a>
              ))}
            </div>
          </div>
        </section>
        <aside className="grid gap-4">
          <div className="rounded-3xl border border-slate-200 bg-white/85 p-5">
            <h2 className="text-lg font-semibold">{locale === "zh" ? "账号摘要" : "Account"}</h2>
            <div className="mt-4 space-y-2 text-sm text-slate-600">
              <p>{data.user.username}</p>
              <p>{data.user.email}</p>
              <p>{data.user.role}</p>
            </div>
            <div className="mt-4 flex flex-wrap gap-2">
              <a className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold" href="/settings/profile">{locale === "zh" ? "资料设置" : "Profile"}</a>
              <a className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold" href="/settings/security">{locale === "zh" ? "安全设置" : "Security"}</a>
            </div>
          </div>
          <div className="rounded-3xl border border-slate-200 bg-white/85 p-5">
            <h2 className="text-lg font-semibold">{locale === "zh" ? "比赛概览" : "Contest overview"}</h2>
            <div className="mt-4 grid gap-3">
              {data.contests.map((item) => (
                <a key={item.id} className="rounded-2xl border border-slate-200 p-4" href={`/contests/${item.slug}`}>
                  <p className="font-semibold text-slate-950">{item.title}</p>
                  <p className="mt-1 text-sm text-slate-500">{item.status}</p>
                </a>
              ))}
            </div>
          </div>
        </aside>
      </div>
    </AppShell>
  );
}
