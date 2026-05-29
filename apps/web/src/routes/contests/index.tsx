import { AppShell } from "../../components/layout/app-shell";
import { useContests } from "../../features/contests/hooks";
import { useLocale } from "../../lib/locale";

const sectionTone = {
  RUNNING: "border-teal-200 bg-teal-50/70",
  UPCOMING: "border-amber-200 bg-amber-50/70",
  ENDED: "border-slate-200 bg-slate-50/70",
} as const;

export function ContestsRoute() {
  const { locale } = useLocale();
  const { data: contests = [], error, isError } = useContests();
  const runningCount = contests.filter((contest) => contest.status === "RUNNING").length;
  const upcomingCount = contests.filter((contest) => contest.status === "UPCOMING").length;
  const endedCount = contests.filter((contest) => contest.status === "ENDED").length;
  const groups = [
    { title: locale === "zh" ? "进行中" : "Running", key: "RUNNING" as const },
    { title: locale === "zh" ? "即将开始" : "Upcoming", key: "UPCOMING" as const },
    { title: locale === "zh" ? "已结束" : "Ended", key: "ENDED" as const },
  ];

  return (
    <AppShell
      title={locale === "zh" ? "比赛" : "Contests"}
      subtitle="luooj"
      action={
        <a className="rounded-full border border-slate-200 bg-white px-4.5 py-2.5 text-sm font-semibold text-slate-700 hover:text-slate-950" href="/">
          {locale === "zh" ? "历史与排名" : "History & Rankings"}
        </a>
      }
    >
      <div className="space-y-4">
        {isError ? (
          <div className="rounded-3xl border border-rose-200 bg-rose-50 px-4 py-4 text-sm text-rose-700">
            {error instanceof Error
              ? error.message
              : locale === "zh"
                ? "比赛列表加载失败。"
                : "Failed to load contests."}
          </div>
        ) : null}
        <section className="grid gap-2.5 md:grid-cols-3">
          {[
            { label: locale === "zh" ? "进行中" : "Running", value: runningCount, tone: "border-teal-200 bg-teal-50/70" },
            { label: locale === "zh" ? "即将开始" : "Upcoming", value: upcomingCount, tone: "border-amber-200 bg-amber-50/70" },
            { label: locale === "zh" ? "已结束" : "Ended", value: endedCount, tone: "border-slate-200 bg-slate-50/90" },
          ].map((item) => (
            <div key={item.label} className={`rounded-2xl border px-4 py-3 ${item.tone}`}>
              <p className="font-mono text-[10px] uppercase tracking-[0.24em] text-slate-500">{item.label}</p>
              <p className="mt-1.5 text-[1.5rem] font-semibold leading-none text-slate-950">{item.value}</p>
              <p className="mt-1 text-xs text-slate-600">{locale === "zh" ? "场比赛" : "events"}</p>
            </div>
          ))}
        </section>

        {groups.map((group) => {
          const items = contests.filter((contest) => contest.status === group.key);
          return (
            <section key={group.key} className="rounded-3xl border border-slate-200/80 bg-white/85 p-4.5">
              <div className="mb-3.5 flex items-end justify-between">
                <div>
                  <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">{group.key}</p>
                  <h2 className="mt-1.5 text-xl font-semibold text-slate-950">{group.title}</h2>
                </div>
                <span className="font-mono text-xs text-slate-400">{locale === "zh" ? `${items.length} 场比赛` : `${items.length} events`}</span>
              </div>
              {items.length === 0 ? (
                <div className="rounded-2xl border border-dashed border-slate-200 p-4">
                  <p className="text-sm font-medium text-slate-900">{locale === "zh" ? `${group.title}比赛为空` : `No ${group.title.toLowerCase()} contests`}</p>
                  <p className="mt-2 text-sm text-slate-500">{locale === "zh" ? "可以先查看其他分区，或者回到题库继续做题。" : "Browse another section or return to the problem set."}</p>
                  <a className="mt-4 inline-flex text-sm font-medium text-slate-900 hover:text-slate-700" href="/problems">
                    {locale === "zh" ? "前往题库" : "Go to problems"}
                  </a>
                </div>
              ) : (
                <div className="grid gap-3.5 lg:grid-cols-2">
                  {items.map((contest) => (
                    <a
                      key={contest.id}
                      className={`block rounded-3xl border p-3.5 transition hover:-translate-y-0.5 hover:shadow-[0_14px_40px_rgba(15,23,42,0.06)] ${sectionTone[contest.status]}`}
                      href={`/contests/${contest.slug}`}
                    >
                      <div className="flex items-start justify-between gap-4">
                        <div>
                          <p className="text-lg font-semibold text-slate-950">{contest.title}</p>
                          <p className="mt-2 max-w-xl text-sm leading-6 text-slate-600">{contest.blurb}</p>
                        </div>
                        <span className="rounded-full bg-white/80 px-2.5 py-1.5 font-mono text-[11px] text-slate-600">{locale === "zh" ? `${contest.problemCount} 题` : `${contest.problemCount} problems`}</span>
                      </div>
                      <div className="mt-3.5 grid gap-2 sm:grid-cols-2 2xl:grid-cols-4">
                        <div className="rounded-xl bg-white/80 px-2.5 py-2">
                          <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-slate-400">{locale === "zh" ? "开始" : "Start"}</p>
                          <p className="mt-1.5 text-sm font-medium text-slate-900">{contest.startsAt}</p>
                        </div>
                        <div className="rounded-xl bg-white/80 px-2.5 py-2">
                          <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-slate-400">{locale === "zh" ? "结束" : "End"}</p>
                          <p className="mt-1.5 text-sm font-medium text-slate-900">{contest.endsAt}</p>
                        </div>
                        <div className="rounded-xl bg-white/80 px-2.5 py-2">
                          <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-slate-400">{locale === "zh" ? "时长" : "Duration"}</p>
                          <p className="mt-1.5 text-sm font-medium text-slate-900">{contest.duration}</p>
                        </div>
                        <div className="rounded-xl bg-white/80 px-2.5 py-2">
                          <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-slate-400">{locale === "zh" ? "人数" : "Players"}</p>
                          <p className="mt-1.5 text-sm font-medium text-slate-900">{contest.participantCount}</p>
                        </div>
                      </div>
                      <div className="mt-3.5 flex items-center justify-between">
                        <span className="text-xs text-slate-500">{contest.endsAt}</span>
                        <span className="rounded-full bg-slate-950 px-3 py-1.5 text-xs font-semibold text-white">
                          {contest.status === "RUNNING"
                            ? locale === "zh"
                              ? "进入比赛"
                              : "Enter Contest"
                            : contest.status === "UPCOMING"
                              ? locale === "zh"
                                ? "查看详情"
                                : "View Details"
                              : locale === "zh"
                                ? "查看榜单"
                                : "View Standings"}
                        </span>
                      </div>
                    </a>
                  ))}
                </div>
              )}
            </section>
          );
        })}
      </div>
    </AppShell>
  );
}
