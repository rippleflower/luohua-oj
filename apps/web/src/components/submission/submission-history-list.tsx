import { webRoutes } from "@oj/shared";
import type { SubmissionSummary } from "../../features/submissions/schema";
import { useLocale } from "../../lib/locale";
import { SubmissionStatusBadge } from "./submission-status-badge";

type SubmissionHistoryListProps = {
  submissions: SubmissionSummary[];
  username: string;
  onUsernameChange: (value: string) => void;
  page: number;
  pageSize: number;
  total: number;
  onPageChange?: (page: number) => void;
  sourceLabel: string;
};

export function SubmissionHistoryList({
  submissions,
  username,
  onUsernameChange,
  page,
  pageSize,
  total,
  onPageChange,
  sourceLabel,
}: SubmissionHistoryListProps) {
  const { locale } = useLocale();
  const totalPages = Math.max(1, Math.ceil(total / Math.max(1, pageSize)));
  const canGoPrev = page > 1;
  const canGoNext = page < totalPages;

  return (
    <section className="rounded-lg border border-slate-200 bg-white p-6">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">{locale === "zh" ? "最近提交" : "Recent"}</h2>
        <a className="text-sm font-medium text-slate-700 hover:underline" href={webRoutes.submissions}>
          {locale === "zh" ? "新建" : "New"}
        </a>
      </div>
      <div className="mt-4 grid gap-2">
        <label className="grid gap-2 text-sm font-medium text-slate-700">
          {locale === "zh" ? "用户名筛选" : "Username Filter"}
          <input
            className="rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-950 outline-none ring-0 transition focus:border-slate-500"
            onChange={(event) => onUsernameChange(event.target.value)}
            placeholder={locale === "zh" ? "按用户名拉取远端提交" : "Fetch remote submissions by username"}
            value={username}
          />
        </label>
        <p className="text-xs text-slate-500">{sourceLabel}</p>
      </div>
      <div className="mt-3 flex items-center justify-between text-xs text-slate-500">
        <span>
          {total === 0 ? (locale === "zh" ? "0 条提交" : "0 submissions") : locale === "zh" ? `${total} 条提交 · 第 ${page} / ${totalPages} 页` : `${total} submissions · page ${page} / ${totalPages}`}
        </span>
        {onPageChange ? (
          <div className="flex items-center gap-2">
            <button
              className="rounded border border-slate-300 px-2 py-1 text-slate-700 disabled:cursor-not-allowed disabled:opacity-50"
              disabled={!canGoPrev}
              onClick={() => onPageChange(page - 1)}
              type="button"
            >
              {locale === "zh" ? "上一页" : "Prev"}
            </button>
            <button
              className="rounded border border-slate-300 px-2 py-1 text-slate-700 disabled:cursor-not-allowed disabled:opacity-50"
              disabled={!canGoNext}
              onClick={() => onPageChange(page + 1)}
              type="button"
            >
              {locale === "zh" ? "下一页" : "Next"}
            </button>
          </div>
        ) : null}
      </div>
      {submissions.length === 0 ? (
        <p className="mt-4 text-sm text-slate-500">{locale === "zh" ? "当前来源下没有提交记录。" : "No submissions available for the current source."}</p>
      ) : (
        <div className="mt-4 overflow-hidden rounded-lg border border-slate-200">
          <table className="min-w-full divide-y divide-slate-200 text-sm">
            <thead className="bg-slate-100 text-left text-xs font-semibold uppercase tracking-wide text-slate-600">
              <tr>
                <th className="px-4 py-3">{locale === "zh" ? "提交" : "Submission"}</th>
                <th className="px-4 py-3">{locale === "zh" ? "题目" : "Problem"}</th>
                <th className="px-4 py-3">{locale === "zh" ? "语言" : "Language"}</th>
                <th className="px-4 py-3">{locale === "zh" ? "状态" : "Status"}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {submissions.map((submission) => (
                <tr key={submission.id} className="hover:bg-slate-50">
                  <td className="px-4 py-4 font-medium text-slate-950">
                    <a className="hover:underline" href={webRoutes.submissionDetail(submission.id)}>
                      {submission.id.slice(0, 8)}
                    </a>
                  </td>
                  <td className="px-4 py-4 text-slate-700">
                    <div className="font-medium text-slate-950">{submission.problem?.title ?? submission.problemId}</div>
                    <div className="text-xs text-slate-500">{submission.problem?.slug ?? submission.problemId}</div>
                  </td>
                  <td className="px-4 py-4 text-slate-700">{submission.language}</td>
                  <td className="px-4 py-4">
                    <SubmissionStatusBadge status={submission.status} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
