import { ApiError } from "../../lib/http-client";
import { useLocale } from "../../lib/locale";
import { AppShell } from "../../components/layout/app-shell";
import { useContestMakeupList } from "../../features/contest-makeup/hooks";
import { webRoutes } from "@oj/shared";

export function ContestMakeupRoute({ slug }: { slug: string }) {
  const { locale } = useLocale();
  const { data, error, isPending } = useContestMakeupList(slug);

  return (
    <AppShell
      title={locale === "zh" ? "比赛后补题清单" : "Contest Makeup List"}
      subtitle="luooj"
      action={
        <a
          className="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700 hover:text-slate-950"
          href={webRoutes.contestDetail(slug)}
        >
          {locale === "zh" ? "返回比赛详情" : "Back to Contest"}
        </a>
      }
    >
      {isPending ? (
        <section className="rounded-3xl border border-slate-200 bg-white p-5 text-sm text-slate-600">
          {locale === "zh" ? "正在生成补题清单..." : "Building makeup list..."}
        </section>
      ) : null}

      {!isPending && error instanceof ApiError && error.status === 409 ? (
        <section className="rounded-3xl border border-amber-200 bg-amber-50 p-5 text-sm text-amber-800">
          {locale === "zh" ? "比赛尚未结束，暂时无法生成补题清单。" : "Contest is not ended yet. Makeup list is unavailable."}
        </section>
      ) : null}

      {!isPending && error instanceof ApiError && error.status === 404 ? (
        <section className="rounded-3xl border border-rose-200 bg-rose-50 p-5 text-sm text-rose-700">
          {locale === "zh" ? "未找到对应比赛。" : "Contest not found."}
        </section>
      ) : null}

      {!isPending && error && (!(error instanceof ApiError) || ![404, 409].includes(error.status)) ? (
        <section className="rounded-3xl border border-rose-200 bg-rose-50 p-5 text-sm text-rose-700">
          {error instanceof Error
            ? error.message
            : locale === "zh"
              ? "补题清单加载失败。"
              : "Failed to load makeup list."}
        </section>
      ) : null}

      {!isPending && !error && data ? (
        <section className="space-y-3">
          {data.items.length === 0 ? (
            <div className="rounded-3xl border border-teal-200 bg-teal-50 p-5 text-sm text-teal-800">
              {locale === "zh" ? "本场比赛题目已全部通过，无需补题。" : "All contest problems are solved. No makeup needed."}
            </div>
          ) : (
            data.items.map((item) => (
              <article key={item.problemId} className="rounded-3xl border border-slate-200 bg-white p-5">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <p className="text-base font-semibold text-slate-950">
                    {item.problemCode}. {item.problemTitle}
                  </p>
                  <span className="rounded-full border border-slate-200 px-3 py-1 font-mono text-xs text-slate-600">
                    {item.category === "ATTEMPTED_UNSOLVED"
                      ? locale === "zh"
                        ? "尝试未过"
                        : "Attempted"
                      : locale === "zh"
                        ? "未尝试"
                        : "Unattempted"}{" "}
                    · {item.difficulty}
                  </span>
                </div>
                <p className="mt-3 text-sm text-slate-700">{item.reasonSummary}</p>
                {item.category === "ATTEMPTED_UNSOLVED" && item.attemptCount ? (
                  <p className="mt-1 text-xs font-mono uppercase tracking-[0.2em] text-slate-500">
                    {locale === "zh" ? `尝试次数 ${item.attemptCount}` : `Attempts ${item.attemptCount}`} · Severity {item.severityRank}
                  </p>
                ) : null}
                <p className="mt-1 text-sm text-slate-600">
                  {locale === "zh" ? "建议动作：" : "Suggested action: "}
                  {item.suggestedAction}
                </p>
                {item.problemSlug ? (
                  <a
                    className="mt-3 inline-flex rounded-full border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-700 hover:text-slate-950"
                    href={`${webRoutes.legacyProblemDetail(item.problemSlug)}?fromMakeup=${encodeURIComponent(slug)}`}
                  >
                    {locale === "zh" ? "打开题目" : "Open Problem"}
                  </a>
                ) : null}
              </article>
            ))
          )}
        </section>
      ) : null}
    </AppShell>
  );
}
