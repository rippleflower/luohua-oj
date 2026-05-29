import { webRoutes } from "@oj/shared";

import { AppShell } from "../../components/layout/app-shell";
import { SubmissionStatusBadge } from "../../components/submission/submission-status-badge";
import { SubmissionResponsePanel } from "../../components/submission/submission-response-panel";
import { listSubmissionHistory } from "../../features/submissions/history";
import { useSubmissionDetail } from "../../features/submissions/hooks";
import { useLocale } from "../../lib/locale";

type SubmissionDetailRouteProps = {
  submissionId: string;
};

export function SubmissionDetailRoute({
  submissionId,
}: SubmissionDetailRouteProps) {
  const { locale } = useLocale();
  const { data: submission, error, isError } = useSubmissionDetail(submissionId);
  const recentSubmissions = listSubmissionHistory()
    .filter((item) => item.id !== submissionId)
    .slice(0, 5);
  const problemTitle = submission?.problem?.title ?? submission?.problemId;
  const problemSlug = submission?.problem?.slug;

  return (
    <AppShell
      title={locale === "zh" ? "提交详情" : "Submission Detail"}
      subtitle="luooj"
      action={
        <div className="flex flex-wrap gap-2">
          <a
            className="rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50"
            href={webRoutes.submissions}
          >
            {locale === "zh" ? "返回提交列表" : "Back to Submissions"}
          </a>
          <a
            className="rounded-md bg-slate-950 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800"
            href={webRoutes.problems}
          >
            {locale === "zh" ? "继续刷题" : "Continue Practice"}
          </a>
        </div>
      }
    >
      <div className="grid gap-6 lg:grid-cols-[minmax(0,2fr)_minmax(320px,1fr)]">
        <div className="space-y-4">
          <SubmissionResponsePanel
            data={submission}
            errorMessage={
              isError
                ? error instanceof Error
                  ? error.message
                  : locale === "zh"
                    ? "提交详情加载失败。"
                    : "Failed to load submission detail."
                : submission
                ? undefined
                : locale === "zh"
                  ? `本地记录里没有找到提交 ${submissionId}。`
                  : `Submission ${submissionId} was not found in local history.`
            }
          />

          {submission ? (
            <section className="rounded-lg border border-slate-200 bg-white p-6">
              <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">
                {locale === "zh" ? "元数据" : "Metadata"}
              </h2>
              <dl className="mt-4 grid gap-4 text-sm sm:grid-cols-2">
                <div>
                  <dt className="text-slate-500">
                    {locale === "zh" ? "用户 ID" : "User ID"}
                  </dt>
                  <dd className="break-all font-medium text-slate-950">
                    {submission.userId}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">
                    {locale === "zh" ? "题目" : "Problem"}
                  </dt>
                  <dd className="break-all font-medium text-slate-950">
                    {problemTitle}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">
                    {locale === "zh" ? "题目 Slug" : "Problem Slug"}
                  </dt>
                  <dd className="break-all font-medium text-slate-950">
                    {problemSlug ?? submission.problemId}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">
                    {locale === "zh" ? "语言" : "Language"}
                  </dt>
                  <dd className="font-medium text-slate-950">
                    {submission.language}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">
                    {locale === "zh" ? "创建时间" : "Created At"}
                  </dt>
                  <dd className="font-medium text-slate-950">
                    {submission.createdAt ??
                      (locale === "zh" ? "未知" : "Unknown")}
                  </dd>
                </div>
              </dl>
              <dl className="mt-4 grid gap-4 text-sm sm:grid-cols-2">
                <div>
                  <dt className="text-slate-500">
                    {locale === "zh" ? "编译摘要" : "Compile Summary"}
                  </dt>
                  <dd className="font-medium text-slate-950">
                    {submission.compileSummary.compileOutput ||
                      (locale === "zh" ? "无" : "None")}
                  </dd>
                </div>
                <div>
                  <dt className="text-slate-500">
                    {locale === "zh" ? "可用产物" : "Artifacts"}
                  </dt>
                  <dd className="font-medium text-slate-950">
                    {submission.artifactAvailability.artifacts.length}
                  </dd>
                </div>
              </dl>
            </section>
          ) : null}
        </div>

        <aside className="space-y-4">
          <section className="rounded-lg border border-slate-200 bg-white p-6">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">
              {locale === "zh" ? "最近提交" : "Recent"}
            </h2>
            {recentSubmissions.length === 0 ? (
              <p className="mt-4 text-sm text-slate-500">
                {locale === "zh"
                  ? "没有其他本地提交。"
                  : "No other local submissions."}
              </p>
            ) : (
              <ul className="mt-4 space-y-3 text-sm">
                {recentSubmissions.map((item) => (
                  <li key={item.id}>
                    <a
                      className="block rounded-md border border-slate-200 px-3 py-2 hover:bg-slate-50"
                      href={webRoutes.submissionDetail(item.id)}
                    >
                      <div className="font-medium text-slate-900">
                        {item.problem?.title ?? item.problemId}
                      </div>
                      <div className="text-xs text-slate-500">
                        {item.problem?.slug ?? item.problemId}
                      </div>
                      <div className="mt-2">
                        <SubmissionStatusBadge status={item.status} />
                      </div>
                    </a>
                  </li>
                ))}
              </ul>
            )}
          </section>
        </aside>
      </div>
    </AppShell>
  );
}
