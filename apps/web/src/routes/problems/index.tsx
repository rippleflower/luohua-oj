import { AppShell } from "../../components/layout/app-shell";
import { ProblemList } from "../../components/problem/problem-list";
import { useProblems } from "../../features/problems/hooks";
import { useLocale } from "../../lib/locale";

export function ProblemsRoute() {
  const { locale } = useLocale();
  const { data: problems = [] } = useProblems();
  const enrichedProblems = problems.map((problem, index) => ({
    ...problem,
    status: (["SOLVED", "ATTEMPTED", "UNSOLVED"][index % 3] as "SOLVED" | "ATTEMPTED" | "UNSOLVED"),
  }));

  return (
    <AppShell
      title={locale === "zh" ? "题库" : "Problems"}
      subtitle="luooj"
      action={
        <a className="rounded-full bg-slate-950 px-4.5 py-2.5 text-sm font-semibold text-white hover:bg-slate-800" href="/submissions">
          {locale === "zh" ? "继续最近提交" : "Resume Latest Submission"}
        </a>
      }
    >
      <div className="grid gap-5 lg:grid-cols-[220px_minmax(0,1fr)] xl:grid-cols-[240px_minmax(0,1fr)]">
        <aside className="grid gap-3.5 rounded-3xl border border-slate-200/80 bg-white/85 p-4.5 sm:grid-cols-2 lg:grid-cols-1">
          <div className="sm:col-span-2 lg:col-span-1">
            <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">{locale === "zh" ? "筛选" : "Filters"}</p>
            <h2 className="mt-1.5 text-lg font-semibold text-slate-950">{locale === "zh" ? "缩小题目范围" : "Refine problems"}</h2>
          </div>
          <div className="rounded-2xl border border-slate-200 bg-slate-50/90 p-3.5">
            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">{locale === "zh" ? "难度" : "Difficulty"}</p>
            <div className="mt-2.5 flex flex-wrap gap-1.5">
              {(locale === "zh" ? ["全部", "简单", "中等", "困难"] : ["All", "Easy", "Medium", "Hard"]).map((item, index) => (
                <button
                  key={item}
                  className={`rounded-full px-3 py-1.5 text-sm ${index === 0 ? "bg-slate-950 text-white" : "bg-white text-slate-600"}`}
                  type="button"
                >
                  {item}
                </button>
              ))}
            </div>
          </div>
          <div className="rounded-2xl border border-slate-200 bg-slate-50/90 p-3.5">
            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">{locale === "zh" ? "状态" : "Status"}</p>
            <div className="mt-2.5 space-y-1.5 text-sm text-slate-600">
              <div className="flex items-center justify-between rounded-xl bg-white px-3 py-1.5"><span>{locale === "zh" ? "未完成" : "Unsolved"}</span><span>14</span></div>
              <div className="flex items-center justify-between rounded-xl bg-white px-3 py-1.5"><span>{locale === "zh" ? "尝试过" : "Attempted"}</span><span>6</span></div>
              <div className="flex items-center justify-between rounded-xl bg-white px-3 py-1.5"><span>{locale === "zh" ? "已通过" : "Solved"}</span><span>23</span></div>
            </div>
          </div>
          <a className="block rounded-2xl border border-slate-200 bg-slate-50/90 p-3.5 hover:border-slate-300 sm:col-span-2 lg:col-span-1" href="/submissions">
            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">{locale === "zh" ? "继续刷题" : "Continue"}</p>
            <p className="mt-2 text-sm font-medium leading-6 text-slate-900">{locale === "zh" ? "回到最近提交，继续调试当前题目。" : "Return to your latest submission and keep iterating."}</p>
          </a>
        </aside>

        <section className="space-y-4">
          <div className="rounded-3xl border border-slate-200/80 bg-white/85 p-4.5">
            <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
              <div>
                <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">{locale === "zh" ? "题目列表" : "Problem Library"}</p>
                <h2 className="mt-1.5 text-lg font-semibold text-slate-950">{locale === "zh" ? "搜索、筛选、直接开做" : "Search, filter, and solve"}</h2>
              </div>
              <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-[minmax(220px,1fr)_140px_140px_160px]">
                <input
                  className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-sm outline-none placeholder:text-slate-400"
                  placeholder={locale === "zh" ? "搜索标题或 slug" : "Search title or slug"}
                />
                <button className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-left text-sm text-slate-600" type="button">
                  {locale === "zh" ? "难度" : "Difficulty"}
                </button>
                <button className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-left text-sm text-slate-600" type="button">
                  {locale === "zh" ? "标签" : "Tags"}
                </button>
                <button className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-left text-sm text-slate-600" type="button">
                  {locale === "zh" ? "状态 / 排序" : "State / Sort"}
                </button>
              </div>
            </div>
          </div>
          <ProblemList problems={enrichedProblems} />
        </section>
      </div>
    </AppShell>
  );
}
