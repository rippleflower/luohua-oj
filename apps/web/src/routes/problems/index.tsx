import { useState } from "react";
import { AppShell } from "../../components/layout/app-shell";
import { ProblemList } from "../../components/problem/problem-list";
import { useProblems } from "../../features/problems/hooks";
import {
  collectProblemTags,
  countProblemStatuses,
  createProblemRows,
  defaultProblemFilterState,
  filterProblemRows,
  type ProblemFilterState,
  type ProblemRowStatus,
} from "../../features/problems/workspace";
import { useLocale } from "../../lib/locale";

export function ProblemsRoute() {
  const { locale } = useLocale();
  const { data: problems = [], error, isError } = useProblems();
  const [filters, setFilters] = useState<ProblemFilterState>(defaultProblemFilterState);
  const enrichedProblems = createProblemRows(problems);
  const problemTags = collectProblemTags(problems);
  const filteredProblems = filterProblemRows(enrichedProblems, filters);
  const counts = countProblemStatuses(enrichedProblems);

  function updateFilters(next: Partial<ProblemFilterState>) {
    setFilters((current) => ({ ...current, ...next }));
  }

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
              {[
                { value: "ALL", label: locale === "zh" ? "全部" : "All" },
                { value: "EASY", label: locale === "zh" ? "简单" : "Easy" },
                { value: "MEDIUM", label: locale === "zh" ? "中等" : "Medium" },
                { value: "HARD", label: locale === "zh" ? "困难" : "Hard" },
              ].map((item) => {
                const active = filters.difficulty === item.value;
                return (
                  <button
                    key={item.value}
                    className={`rounded-full px-3 py-1.5 text-sm transition ${active ? "bg-slate-950 text-white" : "bg-white text-slate-600 hover:bg-slate-100"}`}
                    onClick={() => updateFilters({ difficulty: item.value as ProblemFilterState["difficulty"] })}
                    type="button"
                  >
                    {item.label}
                  </button>
                );
              })}
            </div>
          </div>
          <div className="rounded-2xl border border-slate-200 bg-slate-50/90 p-3.5">
            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">{locale === "zh" ? "状态" : "Status"}</p>
            <div className="mt-2.5 space-y-1.5 text-sm text-slate-600">
              {[
                { value: "ALL", label: locale === "zh" ? "全部状态" : "All states", count: enrichedProblems.length },
                { value: "UNSOLVED", label: locale === "zh" ? "未完成" : "Unsolved", count: counts.UNSOLVED },
                { value: "ATTEMPTED", label: locale === "zh" ? "尝试过" : "Attempted", count: counts.ATTEMPTED },
                { value: "SOLVED", label: locale === "zh" ? "已通过" : "Solved", count: counts.SOLVED },
              ].map((item) => {
                const active = filters.status === item.value;
                return (
                  <button
                    key={item.value}
                    className={`flex w-full items-center justify-between rounded-xl px-3 py-1.5 text-left transition ${active ? "bg-slate-950 text-white" : "bg-white hover:bg-slate-100"}`}
                    onClick={() => updateFilters({ status: item.value as ProblemFilterState["status"] })}
                    type="button"
                  >
                    <span>{item.label}</span>
                    <span>{item.count}</span>
                  </button>
                );
              })}
            </div>
          </div>
          <div className="rounded-2xl border border-slate-200 bg-slate-50/90 p-3.5 sm:col-span-2 lg:col-span-1">
            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">{locale === "zh" ? "标签分类" : "Tags"}</p>
            <div className="mt-2.5 flex flex-wrap gap-1.5">
              <button
                className={`rounded-full px-3 py-1.5 text-sm transition ${filters.tag === "ALL" ? "bg-slate-950 text-white" : "bg-white text-slate-600 hover:bg-slate-100"}`}
                onClick={() => updateFilters({ tag: "ALL" })}
                type="button"
              >
                {locale === "zh" ? "全部标签" : "All tags"}
              </button>
              {problemTags.map((tag) => (
                <button
                  key={tag}
                  className={`rounded-full px-3 py-1.5 text-sm transition ${filters.tag === tag ? "bg-slate-950 text-white" : "bg-white text-slate-600 hover:bg-slate-100"}`}
                  onClick={() => updateFilters({ tag })}
                  type="button"
                >
                  {tag}
                </button>
              ))}
            </div>
          </div>
          <a className="block rounded-2xl border border-slate-200 bg-slate-50/90 p-3.5 hover:border-slate-300 sm:col-span-2 lg:col-span-1" href="/submissions">
            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">{locale === "zh" ? "继续刷题" : "Continue"}</p>
            <p className="mt-2 text-sm font-medium leading-6 text-slate-900">{locale === "zh" ? "回到最近提交，继续调试当前题目。" : "Return to your latest submission and keep iterating."}</p>
          </a>
        </aside>

        <section className="space-y-4">
          {isError ? (
            <div className="rounded-3xl border border-rose-200 bg-rose-50 px-4 py-4 text-sm text-rose-700">
              {error instanceof Error
                ? error.message
                : locale === "zh"
                  ? "题库加载失败。"
                  : "Failed to load problems."}
            </div>
          ) : null}
          <div className="rounded-3xl border border-slate-200/80 bg-white/85 p-4.5">
            <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
              <div>
                <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">{locale === "zh" ? "题目列表" : "Problem Library"}</p>
                <h2 className="mt-1.5 text-lg font-semibold text-slate-950">{locale === "zh" ? "搜索、筛选、直接开做" : "Search, filter, and solve"}</h2>
              </div>
              <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-[minmax(220px,1fr)_140px_140px_160px]">
                <input
                  className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-sm outline-none placeholder:text-slate-400"
                  onChange={(event) => updateFilters({ search: event.target.value })}
                  placeholder={locale === "zh" ? "搜索标题或 slug" : "Search title or slug"}
                  value={filters.search}
                />
                <select
                  className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-left text-sm text-slate-600"
                  onChange={(event) => updateFilters({ difficulty: event.target.value as ProblemFilterState["difficulty"] })}
                  value={filters.difficulty}
                >
                  <option value="ALL">{locale === "zh" ? "全部难度" : "All difficulty"}</option>
                  <option value="EASY">{locale === "zh" ? "简单" : "Easy"}</option>
                  <option value="MEDIUM">{locale === "zh" ? "中等" : "Medium"}</option>
                  <option value="HARD">{locale === "zh" ? "困难" : "Hard"}</option>
                </select>
                <select
                  className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-left text-sm text-slate-600"
                  onChange={(event) => updateFilters({ tag: event.target.value })}
                  value={filters.tag}
                >
                  <option value="ALL">{locale === "zh" ? "全部标签" : "All tags"}</option>
                  {problemTags.map((tag) => (
                    <option key={tag} value={tag}>
                      {tag}
                    </option>
                  ))}
                </select>
                <select
                  className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-2.5 text-left text-sm text-slate-600"
                  onChange={(event) => updateFilters({ status: event.target.value as "ALL" | ProblemRowStatus })}
                  value={filters.status}
                >
                  <option value="ALL">{locale === "zh" ? "全部状态" : "All states"}</option>
                  <option value="UNSOLVED">{locale === "zh" ? "未完成" : "Unsolved"}</option>
                  <option value="ATTEMPTED">{locale === "zh" ? "尝试过" : "Attempted"}</option>
                  <option value="SOLVED">{locale === "zh" ? "已通过" : "Solved"}</option>
                </select>
              </div>
            </div>
          </div>
          <ProblemList
            emptyMessage={locale === "zh" ? "当前筛选条件下没有题目。" : "No problems matched the current filters."}
            onTagClick={(tag) => updateFilters({ tag })}
            problems={filteredProblems}
          />
        </section>
      </div>
    </AppShell>
  );
}
