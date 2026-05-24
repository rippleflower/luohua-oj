import { AppShell } from "../../components/layout/app-shell";
import { useProblem } from "../../features/problems/hooks";
import { useLocale } from "../../lib/locale";

const sectionLabel: Record<string, { zh: string; en: string }> = {
  statement: { zh: "题意", en: "Statement" },
  input: { zh: "输入说明", en: "Input" },
  output: { zh: "输出说明", en: "Output" },
  constraints: { zh: "约束", en: "Constraints" },
};

export function ProblemDetailRoute({ slug }: { slug: string }) {
  const { locale } = useLocale();
  const { data: problem } = useProblem(slug);

  return (
    <AppShell
      title={
        problem?.title ?? (locale === "zh" ? "题目详情" : "Problem Detail")
      }
      subtitle="luooj"
      action={
        <div className="flex flex-wrap gap-3">
          <a
            className="rounded-full bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white hover:bg-slate-800"
            href="/submissions"
          >
            {locale === "zh" ? "去提交" : "Open Submission"}
          </a>
          <a
            className="rounded-full border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 hover:text-slate-950"
            href="/problems"
          >
            {locale === "zh" ? "返回题库" : "Back to Problems"}
          </a>
        </div>
      }
    >
      {!problem ? (
        <div className="rounded-3xl border border-dashed border-slate-200 bg-white/80 p-8 text-sm text-slate-500">
          {locale === "zh" ? "没有找到这道题。" : "Problem not found."}
        </div>
      ) : (
        <div className="grid gap-4 xl:grid-cols-[minmax(0,1.5fr)_320px]">
          <section className="space-y-4">
            <div className="rounded-3xl border border-slate-200/80 bg-white/88 p-5">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">
                    {problem.slug}
                  </p>
                  <h2 className="mt-1.5 text-[1.85rem] font-semibold text-slate-950">
                    {problem.title}
                  </h2>
                </div>
                <div className="flex flex-wrap gap-2">
                  <span className="rounded-full bg-slate-100 px-3 py-1.5 text-xs font-semibold text-slate-700">
                    {locale === "zh" ? "时间" : "Time"}{" "}
                    {problem.limitsJson.timeLimitMs}ms
                  </span>
                  <span className="rounded-full bg-slate-100 px-3 py-1.5 text-xs font-semibold text-slate-700">
                    {locale === "zh" ? "内存" : "Memory"}{" "}
                    {problem.limitsJson.memoryLimitKb}KB
                  </span>
                </div>
              </div>
            </div>

            {problem.statementJson.map((section) => (
              <section
                key={section.section}
                className="rounded-3xl border border-slate-200/80 bg-white/85 p-5"
              >
                <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">
                  {locale === "zh"
                    ? (sectionLabel[section.section]?.zh ?? section.section)
                    : (sectionLabel[section.section]?.en ?? section.section)}
                </p>
                <div className="mt-3 whitespace-pre-wrap text-sm leading-7 text-slate-700">
                  {section.content}
                </div>
              </section>
            ))}

            <section className="rounded-3xl border border-slate-200/80 bg-white/85 p-5">
              <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">
                {locale === "zh" ? "公开样例对象" : "Public Sample Objects"}
              </p>
              <div className="mt-3 grid gap-3 lg:grid-cols-2">
                {problem.samplesJson.map((sample, index) => (
                  <div
                    key={`${sample.inputObjectKey}-${index}`}
                    className="rounded-2xl border border-slate-200 bg-slate-50 p-4"
                  >
                    <p className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">
                      {locale === "zh"
                        ? `样例 ${index + 1}`
                        : `Sample ${index + 1}`}
                    </p>
                    <p className="mt-3 text-xs text-slate-500">
                      {locale === "zh" ? "输入对象" : "Input Object"}
                    </p>
                    <pre className="mt-1 overflow-x-auto whitespace-pre-wrap break-all rounded-xl bg-slate-950 px-3 py-2 text-xs text-slate-100">
                      {sample.inputObjectKey}
                    </pre>
                    <p className="mt-3 text-xs text-slate-500">
                      {locale === "zh" ? "输出对象" : "Output Object"}
                    </p>
                    <pre className="mt-1 overflow-x-auto whitespace-pre-wrap break-all rounded-xl bg-slate-100 px-3 py-2 text-xs text-slate-700">
                      {sample.outputObjectKey}
                    </pre>
                  </div>
                ))}
              </div>
            </section>
          </section>

          <aside className="space-y-4">
            <section className="rounded-3xl border border-slate-200/80 bg-slate-950 p-4.5 text-white">
              <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-teal-200/80">
                {locale === "zh" ? "题目快照" : "Snapshot"}
              </p>
              <div className="mt-3 grid gap-3">
                <div>
                  <p className="text-sm text-slate-300">
                    {locale === "zh" ? "难度" : "Difficulty"}
                  </p>
                  <p className="mt-1 text-lg font-semibold">
                    {problem.difficulty}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-slate-300">
                    {locale === "zh" ? "更新时间" : "Updated At"}
                  </p>
                  <p className="mt-1 text-sm font-medium">
                    {problem.updatedAt}
                  </p>
                </div>
              </div>
            </section>

            <section className="rounded-3xl border border-slate-200/80 bg-white/85 p-4">
              <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">
                {locale === "zh" ? "元数据" : "Metadata"}
              </p>
              <div className="mt-3 space-y-2">
                {Object.entries(problem.metadataJson).map(([key, value]) => (
                  <div
                    key={key}
                    className="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2.5"
                  >
                    <p className="text-xs uppercase tracking-[0.22em] text-slate-400">
                      {key}
                    </p>
                    <p className="mt-1 break-all text-sm font-medium text-slate-900">
                      {typeof value === "string"
                        ? value
                        : JSON.stringify(value)}
                    </p>
                  </div>
                ))}
              </div>
            </section>
          </aside>
        </div>
      )}
    </AppShell>
  );
}
