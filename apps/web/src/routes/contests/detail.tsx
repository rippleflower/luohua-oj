import { AppShell } from "../../components/layout/app-shell";
import { SubmissionStatusBadge } from "../../components/submission/submission-status-badge";
import { useContest } from "../../features/contests/hooks";
import { useLocale } from "../../lib/locale";

const problemTone = {
  SOLVED: "bg-teal-100 text-teal-700",
  ATTEMPTED: "bg-amber-100 text-amber-700",
  LOCKED: "bg-slate-100 text-slate-500",
} as const;

const problemStatusLabel = {
  SOLVED: "已通过",
  ATTEMPTED: "尝试过",
  LOCKED: "未开始",
} as const;

const webRoutes = {
  contests: "/contests",
  legacyProblemDetail: (slug: string) => `/problems/${encodeURIComponent(slug)}`,
  problems: "/problems",
  submissions: "/submissions",
} as const;

export function ContestDetailRoute({ slug }: { slug: string }) {
  const { locale } = useLocale();
  const { data: contest } = useContest(slug);

  return (
    <AppShell
      title={contest?.title ?? (locale === "zh" ? "比赛概览" : "Contest Overview")}
      subtitle="luooj"
      action={
        <a
          className="rounded-full border border-slate-200 bg-white px-4.5 py-2.5 text-sm font-semibold text-slate-700 hover:text-slate-950"
          href={webRoutes.contests}
        >
          {locale === "zh" ? "返回比赛列表" : "Back to Contests"}
        </a>
      }
    >
      {!contest ? (
        <div className="rounded-3xl border border-dashed border-slate-200 bg-white/80 p-8 text-sm text-slate-500">
          {locale === "zh" ? "没有找到这场比赛。" : "Contest not found."}
        </div>
      ) : (
        <div className="grid gap-4 lg:grid-cols-[220px_minmax(0,1fr)] xl:grid-cols-[240px_minmax(0,1.4fr)_280px]">
          <aside className="rounded-3xl border border-slate-200/80 bg-white/85 p-3.5">
            <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">
              {locale === "zh" ? "题目导航" : "Problem Set"}
            </p>
            <div className="mt-3.5 space-y-2">
              {contest.problems.map((problem) =>
                problem.slug ? (
                  <a
                    key={problem.code}
                    className="flex items-start justify-between rounded-2xl border border-slate-200 bg-slate-50/90 px-3 py-2.5 text-left hover:border-slate-300"
                    href={webRoutes.legacyProblemDetail(problem.slug)}
                  >
                    <ProblemSummary locale={locale} problem={problem} />
                  </a>
                ) : (
                  <div
                    key={problem.code}
                    className="flex items-start justify-between rounded-2xl border border-slate-200 bg-slate-50/90 px-3 py-2.5 text-left"
                  >
                    <ProblemSummary locale={locale} problem={problem} />
                  </div>
                ),
              )}
            </div>
          </aside>

          <section className="rounded-3xl border border-slate-200/80 bg-white/88 p-4.5">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">
                  {locale === "zh" ? contest.status : `Status: ${contest.status}`}
                </p>
                <h2 className="mt-1.5 text-[1.75rem] font-semibold text-slate-950">
                  {contest.title}
                </h2>
                <p className="mt-2 text-sm leading-6 text-slate-600">
                  {contest.blurb}
                </p>
              </div>
              <div className="flex flex-wrap gap-2">
                <a
                  className="rounded-full border border-slate-200 bg-white px-3.5 py-2 text-sm font-semibold text-slate-700 hover:text-slate-950"
                  href={webRoutes.problems}
                >
                  {locale === "zh" ? "打开题库" : "Open Problems"}
                </a>
                <a
                  className="rounded-full bg-slate-950 px-3.5 py-2 text-sm font-semibold text-white hover:bg-slate-800"
                  href={webRoutes.submissions}
                >
                  {locale === "zh" ? "查看提交" : "Open Submissions"}
                </a>
              </div>
            </div>

            <div className="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              <MetricCard
                label={locale === "zh" ? "开始" : "Start"}
                value={contest.startsAt}
              />
              <MetricCard
                label={locale === "zh" ? "结束" : "End"}
                value={contest.endsAt}
              />
              <MetricCard
                label={locale === "zh" ? "时长" : "Duration"}
                value={contest.duration}
              />
              <MetricCard
                label={locale === "zh" ? "题目数" : "Problems"}
                value={`${contest.problemCount}`}
              />
            </div>

            <div className="mt-4 grid gap-3 rounded-3xl border border-slate-200 bg-white p-4 lg:grid-cols-2">
              <MetricCard
                label={locale === "zh" ? "参赛人数" : "Participants"}
                value={`${contest.participantCount}`}
              />
              <MetricCard
                label={locale === "zh" ? "当前排名摘要" : "Ranking Summary"}
                value={contest.rankSummary}
              />
            </div>
          </section>

          <aside className="grid gap-3 sm:grid-cols-2 lg:col-span-2 xl:col-span-1 xl:grid-cols-1">
            <section className="rounded-3xl border border-slate-200/80 bg-slate-950 p-4.5 text-white sm:col-span-2 xl:col-span-1">
              <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-teal-200/80">
                {locale === "zh" ? "比赛状态" : "Contest Status"}
              </p>
              <div className="mt-3.5 grid gap-2.5">
                <div>
                  <p className="text-[1.75rem] font-semibold leading-none">
                    {contest.remaining}
                  </p>
                  <p className="mt-1 text-sm text-slate-300">
                    {locale === "zh" ? "剩余时间" : "Remaining time"}
                  </p>
                </div>
                <p className="text-sm leading-6 text-slate-300">
                  {locale === "zh"
                    ? "比赛页仅展示真实概览数据，题目详情请从左侧题目导航进入。"
                    : "This page only shows live contest overview data. Open problems from the left navigation."}
                </p>
              </div>
            </section>
            <section className="rounded-3xl border border-slate-200/80 bg-white/85 p-4">
              <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">
                {locale === "zh" ? "最近提交" : "Recent submissions"}
              </p>
              <div className="mt-3 space-y-2">
                {contest.recentSubmissions.length === 0 ? (
                  <div className="rounded-2xl border border-dashed border-slate-200 px-3.5 py-3.5 text-sm text-slate-500">
                    {locale === "zh"
                      ? "比赛尚未开始，当前没有提交记录。"
                      : "No contest submissions yet."}
                  </div>
                ) : (
                  contest.recentSubmissions.map((submission) => (
                    <div
                      key={submission.id}
                      className="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2.5"
                    >
                      <div className="flex items-center justify-between gap-4">
                        <p className="text-sm font-semibold text-slate-900">
                          {locale === "zh"
                            ? `题目 ${submission.problemCode}`
                            : `Problem ${submission.problemCode}`}
                        </p>
                        <span className="font-mono text-xs text-slate-400">
                          {submission.at}
                        </span>
                      </div>
                      <div className="mt-1">
                        <SubmissionStatusBadge status={submission.status} />
                      </div>
                    </div>
                  ))
                )}
              </div>
            </section>
          </aside>
        </div>
      )}
    </AppShell>
  );
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-slate-50 p-3.5">
      <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-slate-400">
        {label}
      </p>
      <p className="mt-1.5 text-sm font-medium text-slate-900">{value}</p>
    </div>
  );
}

function ProblemSummary({
  locale,
  problem,
}: {
  locale: string;
  problem: {
    code: string;
    title: string;
    status: "LOCKED" | "ATTEMPTED" | "SOLVED";
    firstSolve?: string;
  };
}) {
  return (
    <>
      <div>
        <p className="font-mono text-xs uppercase tracking-[0.22em] text-slate-400">
          {problem.code}
        </p>
        <p className="mt-1.5 text-sm font-semibold text-slate-950">
          {problem.title}
        </p>
        {problem.firstSolve ? (
          <p className="mt-0.5 text-xs text-slate-500">
            {locale === "zh"
              ? `首次通过 ${problem.firstSolve}`
              : `First solved at ${problem.firstSolve}`}
          </p>
        ) : null}
      </div>
      <span
        className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${problemTone[problem.status]}`}
      >
        {locale === "zh" ? problemStatusLabel[problem.status] : problem.status}
      </span>
    </>
  );
}
