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

export function ContestDetailRoute({ slug }: { slug: string }) {
  const { locale } = useLocale();
  const { data: contest } = useContest(slug);
  const currentProblem = contest?.problems.find((problem) => problem.status !== "LOCKED") ?? contest?.problems[0];

  return (
    <AppShell
      title={contest?.title ?? (locale === "zh" ? "比赛工作台" : "Contest Workspace")}
      subtitle="luooj"
      action={
        <a className="rounded-full border border-slate-200 bg-white px-4.5 py-2.5 text-sm font-semibold text-slate-700 hover:text-slate-950" href="/contests">
          {locale === "zh" ? "返回比赛列表" : "Back to Contests"}
        </a>
      }
    >
      {!contest ? (
        <div className="rounded-3xl border border-dashed border-slate-200 bg-white/80 p-8 text-sm text-slate-500">{locale === "zh" ? "没有找到这场比赛。" : "Contest not found."}</div>
      ) : (
        <div className="grid gap-4 lg:grid-cols-[200px_minmax(0,1fr)] xl:grid-cols-[220px_minmax(0,1.48fr)_260px]">
          <aside className="rounded-3xl border border-slate-200/80 bg-white/85 p-3.5">
            <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">{locale === "zh" ? "题目导航" : "Problem Set"}</p>
            <div className="mt-3.5 space-y-2">
              {contest.problems.map((problem) => (
                <button
                  key={problem.code}
                  className="flex w-full items-start justify-between rounded-2xl border border-slate-200 bg-slate-50/90 px-3 py-2.5 text-left hover:border-slate-300"
                  type="button"
                >
                  <div>
                    <p className="font-mono text-xs uppercase tracking-[0.22em] text-slate-400">{problem.code}</p>
                    <p className="mt-1.5 text-sm font-semibold text-slate-950">{problem.title}</p>
                    {problem.firstSolve ? <p className="mt-0.5 text-xs text-slate-500">{locale === "zh" ? `首次通过 ${problem.firstSolve}` : `First solved at ${problem.firstSolve}`}</p> : null}
                  </div>
                  <span className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${problemTone[problem.status]}`}>
                    {locale === "zh" ? problemStatusLabel[problem.status] : problem.status}
                  </span>
                </button>
              ))}
            </div>
          </aside>

          <section className="rounded-3xl border border-slate-200/80 bg-white/88 p-4.5">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">{locale === "zh" ? `${contest.startsAt} · ${contest.duration}` : `${contest.startsAt} · ${contest.duration}`}</p>
                <h2 className="mt-1.5 text-[1.75rem] font-semibold text-slate-950">{currentProblem?.code}. {currentProblem?.title}</h2>
              </div>
              <a className="rounded-full bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white hover:bg-slate-800" href="/submissions">
                {locale === "zh" ? "去提交" : "Open Submission"}
              </a>
            </div>
            <div className="mt-5 space-y-3">
              <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-slate-400">{locale === "zh" ? "题意" : "Statement"}</p>
                <p className="mt-3 text-sm leading-7 text-slate-600">{contest.blurb}</p>
              </div>
              <div className="grid gap-3 sm:grid-cols-2">
                <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                  <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-slate-400">{locale === "zh" ? "输入说明" : "Input"}</p>
                  <p className="mt-3 text-sm leading-7 text-slate-600">
                    {locale === "zh"
                      ? "第一行给出测试组数。每组包含一段操作序列和若干查询，输入规模允许线性或接近线性的单次扫描。"
                      : "The first line contains the number of test cases. Each case includes an operation stream and several queries."}
                  </p>
                </div>
                <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                  <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-slate-400">{locale === "zh" ? "输出说明" : "Output"}</p>
                  <p className="mt-3 text-sm leading-7 text-slate-600">
                    {locale === "zh"
                      ? "对每组查询输出稳定、可校验的结果格式。优先保证边界情况和多组数据之间的状态隔离。"
                      : "Print one deterministic answer per query while keeping per-case state isolated."}
                  </p>
                </div>
              </div>
              <div className="rounded-3xl border border-slate-200 bg-white p-4">
                <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-400">{locale === "zh" ? "样例" : "Sample"}</p>
                <div className="mt-3 grid gap-3 lg:grid-cols-2">
                  <div className="rounded-2xl bg-slate-950 p-4 text-sm text-slate-100">
                    <p className="font-mono text-xs uppercase tracking-[0.22em] text-slate-400">{locale === "zh" ? "样例输入" : "Sample Input"}</p>
                    <pre className="mt-3 whitespace-pre-wrap font-mono text-xs leading-6 text-slate-200">2{"\n"}5 3{"\n"}1 2 3 4 5{"\n"}2 4</pre>
                  </div>
                  <div className="rounded-2xl bg-slate-100 p-4 text-sm text-slate-900">
                    <p className="font-mono text-xs uppercase tracking-[0.22em] text-slate-500">{locale === "zh" ? "样例输出" : "Sample Output"}</p>
                    <pre className="mt-3 whitespace-pre-wrap font-mono text-xs leading-6 text-slate-700">7{"\n"}9</pre>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <aside className="grid gap-3 sm:grid-cols-2 lg:col-span-2 xl:col-span-1 xl:grid-cols-1">
            <section className="rounded-3xl border border-slate-200/80 bg-slate-950 p-4.5 text-white sm:col-span-2 xl:col-span-1">
              <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-teal-200/80">{locale === "zh" ? "比赛状态" : "Contest Status"}</p>
              <div className="mt-3.5 grid gap-2.5">
                <div>
                  <p className="text-[1.75rem] font-semibold leading-none">{contest.remaining}</p>
                  <p className="mt-1 text-sm text-slate-300">{locale === "zh" ? "剩余时间" : "Remaining time"}</p>
                </div>
                <div>
                  <p className="text-base font-semibold">{contest.rankSummary}</p>
                  <p className="mt-1 text-sm text-slate-300">{locale === "zh" ? "当前排名摘要" : "Current standing snapshot"}</p>
                </div>
              </div>
            </section>
            <section className="rounded-3xl border border-slate-200/80 bg-white/85 p-4">
              <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">{locale === "zh" ? "最近提交" : "Recent submissions"}</p>
              <div className="mt-3 space-y-2">
                {contest.recentSubmissions.length === 0 ? (
                  <div className="rounded-2xl border border-dashed border-slate-200 px-3.5 py-3.5 text-sm text-slate-500">
                    {locale === "zh" ? "比赛尚未开始，当前没有提交记录。" : "No contest submissions yet."}
                  </div>
                ) : (
                  contest.recentSubmissions.map((submission) => (
                    <div key={submission.id} className="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2.5">
                      <div className="flex items-center justify-between gap-4">
                        <p className="text-sm font-semibold text-slate-900">{locale === "zh" ? `题目 ${submission.problemCode}` : `Problem ${submission.problemCode}`}</p>
                        <span className="font-mono text-xs text-slate-400">{submission.at}</span>
                      </div>
                      <div className="mt-1">
                        <SubmissionStatusBadge status={submission.status} />
                      </div>
                    </div>
                  ))
                )}
              </div>
            </section>
            <a className="block rounded-3xl border border-slate-200/80 bg-white/85 p-4 hover:border-slate-300" href="/contests">
              <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">{locale === "zh" ? "榜单" : "Standings"}</p>
              <p className="mt-2 text-base font-semibold text-slate-950">{locale === "zh" ? "查看比赛榜单" : "Open contest rankings"}</p>
              <p className="mt-1.5 text-sm leading-6 text-slate-600">{locale === "zh" ? "当前排名、罚时和解题情况统一从榜单查看，工作台只保留一个清晰入口。" : "Use one clear action to jump to standings."}</p>
            </a>
          </aside>
        </div>
      )}
    </AppShell>
  );
}
