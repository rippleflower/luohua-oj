import { AppShell } from "../components/layout/app-shell";
import { SubmissionStatusBadge } from "../components/submission/submission-status-badge";
import { useContests } from "../features/contests/hooks";
import { listSubmissionHistory } from "../features/submissions/history";
import { useProblems } from "../features/problems/hooks";
import { useLocale } from "../lib/locale";

export function HomeRoute() {
  const { locale } = useLocale();
  const { data: problems = [] } = useProblems();
  const { data: contests = [] } = useContests();
  const recentSubmissions = listSubmissionHistory().slice(0, 3);
  const runningContests = contests.filter((contest) => contest.status === "RUNNING");
  const upcomingContests = contests.filter((contest) => contest.status === "UPCOMING");
  const popularTags = [...new Set(problems.flatMap((problem) => problem.tags))].slice(0, 5);

  return (
    <AppShell
      title={locale === "zh" ? "面向刷题与比赛的在线判题台" : "A cleaner online judge for practice and contests."}
      subtitle="luooj"
      action={
        <div className="flex flex-wrap gap-3">
          <a className="rounded-full bg-slate-950 px-4.5 py-2.5 text-sm font-semibold text-white hover:bg-slate-800" href="/problems">
            {locale === "zh" ? "开始做题" : "Start Solving"}
          </a>
          <a className="rounded-full border border-slate-200 bg-white px-4.5 py-2.5 text-sm font-semibold text-slate-700 hover:text-slate-950" href="/contests">
            {locale === "zh" ? "查看比赛" : "View Contests"}
          </a>
        </div>
      }
    >
      <div className="grid gap-5 md:grid-cols-[minmax(0,1.55fr)_minmax(240px,0.88fr)] xl:grid-cols-[1.8fr_minmax(260px,0.82fr)]">
        <section className="rounded-3xl border border-slate-200/80 bg-white/85 p-6 shadow-[0_18px_60px_rgba(15,23,42,0.05)]">
          <div className="flex flex-wrap items-center gap-3 font-mono text-[11px] uppercase tracking-[0.3em] text-slate-500">
            <span>{locale === "zh" ? "题库" : "Problems"}</span>
            <span className="h-1 w-1 rounded-full bg-slate-300" />
            <span>{locale === "zh" ? "比赛" : "Contests"}</span>
            <span className="h-1 w-1 rounded-full bg-slate-300" />
            <span>{locale === "zh" ? "提交反馈" : "Submissions"}</span>
          </div>
          <h2 className="mt-5 max-w-3xl text-[2rem] font-semibold leading-tight text-slate-950 sm:text-[2.75rem]">
            {locale === "zh" ? "把做题、比赛和提交反馈放进同一个清晰工作流。" : "Keep solving, contesting, and submission feedback in one clean workflow."}
          </h2>
          <p className="mt-5 max-w-2xl text-base leading-7 text-slate-600">
            {locale === "zh"
              ? "题库负责持续刷题，比赛页负责倒计时和排名节奏，提交页负责快速反馈。首页只保留最短路径，直接把你送到下一次 Accepted。"
              : "Use the problem library for steady practice, contests for timed work, and the submissions page for fast feedback."}
          </p>
          <div className="mt-7 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
            <a className="rounded-2xl border border-slate-200 bg-slate-50/90 p-3.5 hover:border-slate-300" href="/problems">
              <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-slate-500">{locale === "zh" ? "Problem Set" : "Problem Set"}</p>
              <p className="mt-2.5 text-base font-semibold text-slate-950">{locale === "zh" ? "进入题库" : "Open Problems"}</p>
              <p className="mt-1.5 text-sm leading-6 text-slate-600">{locale === "zh" ? "按难度、标签和完成状态筛选，直接在表格里决定下一题。" : "Filter by difficulty, tags, and solving state in one table view."}</p>
              <p className="mt-3 text-sm font-medium text-slate-900">{locale === "zh" ? "查看题目" : "Browse problems"}</p>
            </a>
            <a className="rounded-2xl border border-slate-200 bg-slate-50/90 p-3.5 hover:border-slate-300" href="/contests">
              <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-slate-500">{locale === "zh" ? "Contest Hub" : "Contest Hub"}</p>
              <p className="mt-2.5 text-base font-semibold text-slate-950">{locale === "zh" ? "查看比赛" : "Open Contests"}</p>
              <p className="mt-1.5 text-sm leading-6 text-slate-600">{locale === "zh" ? "从比赛大厅进入工作台，优先看到倒计时、题目导航和最近提交。" : "Open the contest workspace with countdown, problem nav, and recent runs."}</p>
              <p className="mt-3 text-sm font-medium text-slate-900">{locale === "zh" ? "进入大厅" : "Open hub"}</p>
            </a>
            <a className="rounded-2xl border border-slate-200 bg-slate-50/90 p-3.5 hover:border-slate-300" href="/submissions">
              <p className="font-mono text-[11px] uppercase tracking-[0.24em] text-slate-500">{locale === "zh" ? "Submission Feed" : "Submission Feed"}</p>
              <p className="mt-2.5 text-base font-semibold text-slate-950">{locale === "zh" ? "最近提交" : "Recent Submissions"}</p>
              <p className="mt-1.5 text-sm leading-6 text-slate-600">{locale === "zh" ? "把题目标题、状态变化和提交记录放在一起，方便连续迭代。" : "Keep problem titles and feedback close while iterating."}</p>
              <p className="mt-3 text-sm font-medium text-slate-900">{locale === "zh" ? "打开提交台" : "Open submissions"}</p>
            </a>
          </div>
        </section>

        <aside className="grid gap-3 sm:grid-cols-2 md:grid-cols-1">
          <div className="rounded-3xl border border-teal-900/10 bg-slate-950 p-4.5 text-white shadow-[0_16px_56px_rgba(15,23,42,0.16)]">
            <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-teal-200/80">{locale === "zh" ? "今日概览" : "Today Overview"}</p>
            <div className="mt-3 divide-y divide-white/10 sm:grid sm:grid-cols-3 sm:divide-x sm:divide-y-0 lg:block lg:divide-x-0 lg:divide-y">
              <div className="py-2 sm:px-0 sm:py-0 lg:py-2">
                <p className="text-[1.5rem] font-semibold leading-none">{problems.length}</p>
                <p className="mt-1 text-xs text-slate-300">{locale === "zh" ? "当前题目数" : "Problems loaded today"}</p>
              </div>
              <div className="py-2 sm:px-3 sm:py-0 lg:px-0 lg:py-2">
                <p className="text-[1.5rem] font-semibold leading-none">{runningContests.length}</p>
                <p className="mt-1 text-xs text-slate-300">{locale === "zh" ? "进行中比赛" : "Contests in flight"}</p>
              </div>
              <div className="py-2 sm:px-3 sm:py-0 lg:px-0 lg:py-2">
                <p className="text-[1.5rem] font-semibold leading-none">{recentSubmissions.length}</p>
                <p className="mt-1 text-xs text-slate-300">{locale === "zh" ? "本地最近提交" : "Recent local submissions"}</p>
              </div>
            </div>
          </div>

          <div className="rounded-3xl border border-slate-200/80 bg-white/85 p-4">
            <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">{locale === "zh" ? "即将开始" : "Upcoming"}</p>
            <div className="mt-3 space-y-2.5">
              {upcomingContests.slice(0, 2).map((contest) => (
                <a key={contest.id} className="block rounded-2xl border border-slate-200 p-3 hover:border-slate-300" href={`/contests/${contest.slug}`}>
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <p className="text-sm font-semibold text-slate-950">{contest.title}</p>
                      <p className="mt-1 text-xs text-slate-500">{contest.startsAt}</p>
                    </div>
                    <span className="rounded-full bg-amber-50 px-2.5 py-1 text-[11px] font-semibold text-amber-700">{contest.duration}</span>
                  </div>
                </a>
              ))}
            </div>
          </div>
        </aside>
      </div>

      <div className="mt-7 grid gap-4 md:grid-cols-[minmax(0,1.12fr)_minmax(260px,0.88fr)] xl:grid-cols-[1.4fr_1fr]">
        <section className="rounded-3xl border border-slate-200/80 bg-white/80 p-4.5">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">{locale === "zh" ? "题目领域" : "Problem Field"}</p>
              <h3 className="mt-1.5 text-lg font-semibold text-slate-950">{locale === "zh" ? "题型分布" : "Topic pressure map"}</h3>
            </div>
            <a className="text-sm font-medium text-slate-600 hover:text-slate-950" href="/problems">
              {locale === "zh" ? "查看全部" : "Browse all"}
            </a>
          </div>
          <div className="mt-4 flex flex-wrap gap-2.5">
            {popularTags.map((tag, index) => (
              <span
                key={tag}
                className="rounded-full border border-slate-200 px-3 py-1.5 text-sm text-slate-700"
                style={{ background: `rgba(15, 118, 110, ${0.08 + index * 0.03})` }}
              >
                {tag}
              </span>
            ))}
          </div>
          <div className="mt-5 grid gap-2.5 sm:grid-cols-3">
            {[
              { label: locale === "zh" ? "简单" : "Easy", value: problems.filter((problem) => problem.difficulty === "EASY").length, tone: "bg-emerald-100 text-emerald-700" },
              { label: locale === "zh" ? "中等" : "Medium", value: problems.filter((problem) => problem.difficulty === "MEDIUM").length, tone: "bg-amber-100 text-amber-700" },
              { label: locale === "zh" ? "困难" : "Hard", value: problems.filter((problem) => problem.difficulty === "HARD").length, tone: "bg-rose-100 text-rose-700" },
            ].map((item) => (
              <div key={item.label} className="rounded-2xl border border-slate-200 bg-slate-50 p-3">
                <p className={`inline-flex rounded-full px-2.5 py-1 text-xs font-semibold ${item.tone}`}>{item.label}</p>
                <p className="mt-2.5 text-[1.5rem] font-semibold leading-none text-slate-950">{item.value}</p>
              </div>
            ))}
          </div>
        </section>

        <section className="rounded-3xl border border-slate-200/80 bg-white/80 p-4.5">
          <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">{locale === "zh" ? "近期动态" : "Recent activity"}</p>
          <h3 className="mt-1.5 text-lg font-semibold text-slate-950">{locale === "zh" ? "最近提交" : "Recent submissions"}</h3>
          <div className="mt-4 space-y-3">
            {recentSubmissions.length === 0 ? (
              <p className="rounded-2xl border border-dashed border-slate-200 p-3.5 text-sm leading-6 text-slate-500">
                {locale === "zh" ? "还没有本地提交。先去提交页创建一条记录，这里会显示最新反馈。" : "No local submissions yet. Create one to populate this feed."}
              </p>
            ) : (
              recentSubmissions.map((item) => (
                <a key={item.id} className="block rounded-2xl border border-slate-200 p-3 hover:border-slate-300" href={`/submissions/${item.id}`}>
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <p className="font-semibold text-slate-950">{item.problem?.title ?? item.problemId}</p>
                      <p className="mt-1 text-xs text-slate-500">{item.problem?.slug ?? item.problemId}</p>
                      <div className="mt-2">
                        <SubmissionStatusBadge status={item.status} />
                      </div>
                    </div>
                    <span className="font-mono text-xs text-slate-400">{item.id.slice(0, 8)}</span>
                  </div>
                </a>
              ))
            )}
          </div>
        </section>
      </div>
    </AppShell>
  );
}
